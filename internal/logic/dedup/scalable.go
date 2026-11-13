package dedup

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"

	bloomfilter "github.com/bits-and-blooms/bloom/v3"
)

const (
	sbfMagic          = uint32(0x53424631) // "SBF1" 序列化格式标识
	sbfGrowthFactor   = 2                  // 新层容量 = 上一层 × 2
	sbfTighteningRate = 0.5                // 新层误判率 = 上一层 × 0.5
)

// scalableFilter 可自动扩容的 Bloom Filter (Scalable Bloom Filter)
//
// 当前层写满后追加一个「容量×2, 误判率×0.5」的新层, 查询时逐层查。
// 总误判率收敛于 baseP × Σ(0.5^i) = 2×baseP, 与写入量无关 —
// 解决了固定容量 filter 超量后误判率飙升 (新设备被静默漏写 CSV) 的问题。
//
// 非并发安全, 由调用方 (guardedFilter) 加锁。
type scalableFilter struct {
	layers []*bloomfilter.BloomFilter
	counts []uint64 // 每层已写入元素数, 用于触发扩容
	caps   []uint64 // 每层容量
	baseP  float64
}

// newScalableFilter 构造, baseN/baseP 为第一层的容量与误判率
func newScalableFilter(baseN uint, baseP float64) *scalableFilter {
	sf := &scalableFilter{baseP: baseP}
	sf.appendLayer(uint64(baseN), baseP)
	return sf
}

func (sf *scalableFilter) appendLayer(n uint64, p float64) {
	if n == 0 {
		n = 1
	}
	if n > math.MaxUint {
		n = math.MaxUint // 32 位平台保护 (实际部署为 64 位, 不会触发)
	}
	sf.layers = append(sf.layers, bloomfilter.NewWithEstimates(uint(n), p))
	sf.counts = append(sf.counts, 0)
	sf.caps = append(sf.caps, n)
}

// layerP 第 i 层 (0-based) 的设计误判率
func (sf *scalableFilter) layerP(i int) float64 {
	return sf.baseP * math.Pow(sbfTighteningRate, float64(i))
}

// TestAndAdd 查询并写入: 任一层已存在返回 true (不写入);
// 否则写入最后一层, 写满则追加新层
func (sf *scalableFilter) TestAndAdd(value string) (existed bool) {
	data := []byte(value)
	for _, bf := range sf.layers {
		if bf.Test(data) {
			return true
		}
	}
	last := len(sf.layers) - 1
	sf.layers[last].Add(data)
	sf.counts[last]++
	if sf.counts[last] >= sf.caps[last] {
		sf.appendLayer(sf.caps[last]*sbfGrowthFactor, sf.layerP(last+1))
	}
	return false
}

// Layers 当前层数 (用于扩容日志/观测)
func (sf *scalableFilter) Layers() int {
	return len(sf.layers)
}

// Count 已写入元素总数 (估算)
func (sf *scalableFilter) Count() uint64 {
	var total uint64
	for _, c := range sf.counts {
		total += c
	}
	return total
}

// writeTo 序列化: magic + baseP + 层数 + 每层 (count + cap + bloom 二进制)
func (sf *scalableFilter) writeTo(w io.Writer) error {
	bw := bufio.NewWriter(w)
	if err := binary.Write(bw, binary.BigEndian, sbfMagic); err != nil {
		return err
	}
	if err := binary.Write(bw, binary.BigEndian, sf.baseP); err != nil {
		return err
	}
	if err := binary.Write(bw, binary.BigEndian, uint32(len(sf.layers))); err != nil {
		return err
	}
	for i, bf := range sf.layers {
		if err := binary.Write(bw, binary.BigEndian, sf.counts[i]); err != nil {
			return err
		}
		if err := binary.Write(bw, binary.BigEndian, sf.caps[i]); err != nil {
			return err
		}
		if _, err := bf.WriteTo(bw); err != nil {
			return err
		}
	}
	return bw.Flush()
}

// readScalableFilter 反序列化。兼容旧版"单个裸 BloomFilter"格式:
// 旧文件作为第 0 层载入, 容量从 m/k 反推, 已写入数用 ApproximatedSize 估算,
// 旧层写满后照常扩容。
func readScalableFilter(r io.Reader, defaultP float64) (*scalableFilter, error) {
	raw, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	if len(raw) >= 4 && binary.BigEndian.Uint32(raw[:4]) == sbfMagic {
		return parseSBF(bytes.NewReader(raw[4:]))
	}
	return parseLegacy(raw, defaultP)
}

// safeReadBloom 反序列化单个 BloomFilter
// bitset.ReadFrom 对损坏数据会 panic (makeslice: len out of range), 必须 recover 转为 error
func safeReadBloom(r io.Reader) (bf *bloomfilter.BloomFilter, err error) {
	defer func() {
		if rec := recover(); rec != nil {
			bf, err = nil, fmt.Errorf("bloom deserialize panic: %v", rec)
		}
	}()
	bf = &bloomfilter.BloomFilter{}
	if _, err = bf.ReadFrom(r); err != nil {
		return nil, err
	}
	return bf, nil
}

func parseSBF(br *bytes.Reader) (*scalableFilter, error) {
	sf := &scalableFilter{}
	if err := binary.Read(br, binary.BigEndian, &sf.baseP); err != nil {
		return nil, err
	}
	// baseP 用于扩容时计算新层误判率, 病态值会让 NewWithEstimates 产生畸形参数
	if sf.baseP <= 0 || sf.baseP >= 1 || math.IsNaN(sf.baseP) {
		return nil, fmt.Errorf("invalid sbf baseP: %v", sf.baseP)
	}
	var numLayers uint32
	if err := binary.Read(br, binary.BigEndian, &numLayers); err != nil {
		return nil, err
	}
	if numLayers == 0 || numLayers > 64 {
		return nil, fmt.Errorf("invalid sbf layer count: %d", numLayers)
	}
	for i := uint32(0); i < numLayers; i++ {
		var count, capacity uint64
		if err := binary.Read(br, binary.BigEndian, &count); err != nil {
			return nil, err
		}
		if err := binary.Read(br, binary.BigEndian, &capacity); err != nil {
			return nil, err
		}
		// capacity=0 会让后续扩容尺寸 (cap×2) 永远为 0
		if capacity == 0 || capacity > math.MaxUint64/sbfGrowthFactor {
			return nil, fmt.Errorf("invalid sbf layer %d capacity: %d", i, capacity)
		}
		bf, err := safeReadBloom(br)
		if err != nil {
			return nil, err
		}
		sf.layers = append(sf.layers, bf)
		sf.counts = append(sf.counts, count)
		sf.caps = append(sf.caps, capacity)
	}
	return sf, nil
}

func parseLegacy(raw []byte, defaultP float64) (*scalableFilter, error) {
	bf, err := safeReadBloom(bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	if bf.K() == 0 {
		return nil, fmt.Errorf("invalid legacy bloom filter")
	}
	// 容量反推: k = m/n × ln2  →  n = m × ln2 / k
	capacity := uint64(float64(bf.Cap()) * math.Ln2 / float64(bf.K()))
	if capacity == 0 {
		capacity = 1
	}
	return &scalableFilter{
		layers: []*bloomfilter.BloomFilter{bf},
		counts: []uint64{uint64(bf.ApproximatedSize())},
		caps:   []uint64{capacity},
		baseP:  defaultP,
	}, nil
}
