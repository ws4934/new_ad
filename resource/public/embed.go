// Package public 把前端静态文件编译进二进制, 在线更新只需替换单个可执行文件。
package public

import "embed"

//go:embed index.html js
var FS embed.FS
