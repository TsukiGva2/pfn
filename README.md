# pfn

this is the go package for pfn, i made it a package for using it with the [pfn-play](https://github.com/TsukiGva2/pfn-play) repo, which is a version of the language running on the web.

### using this code

if you, for any weird reason, need to use pfn in your go code, just type

    $ go get github.com/TsukiGva2/pfn

in your project's folder, and try transpiling a simple script with the following code:

```go
import "github.com/TsukiGva2/pfn"

func main() {
  pfn.Run(".f(|x|->0)", true) // true is for printing the output to the console
}
```
