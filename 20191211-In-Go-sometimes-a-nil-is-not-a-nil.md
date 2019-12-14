# In Go, sometimes a nil is not a nil

今天，我遇到了一个 Go 常见问题解答。首先，作为一个小小的 Go 测验，看看您是否可以在 Go 操场上运行该程序之前就知道该程序应打印的内容（我将程序放在侧边栏中，以防万一它在操场上消失了）。该程序的核心是以下代码：

```golang
type fake struct { io.Writer }

func fred (logger io.Writer) {
   if logger != nil {
      logger.Write([]byte("..."))
   }
}

func main() {
   var lp *fake
   fred(nil)
   fred(lp)
}
```

由于 Go 变量是使用零值显式创建的，在 lp 等指针为零的情况下，您希望该代码运行（但不执行任何操作）。实际上，它在对 fred() 的第二次调用时崩溃。发生的情况是，有时在 Go 中，最初以 nil 值开头，如果直接打印它看起来像 nil 值，实际上并不是 nil 值。简而言之，Go 可以区分 nil 接口值和转换为接口的具体类型的 nil 值。只有前者确实为 nil，因此可以将其与纯ni​​l进行比较，就像 fred() 试图在此处这样做。

（作为一个必然结果，可以使用 nil f 值调用（f *fake）上的具体方法。它可以是nil指针，但是它是类型化的 nil 指针，因此可以有方法。甚至可以有方法通过接口，如此处所示。）

对于我发现的情况，解决此问题的方法是更改​​设置过程。真正的程序有条件地设置了伪造，例如：

```golang
var l *sLogger

if smtplog != nil {
    l = &sLogger
    l.prefix = logpref
    l.writer = bufio.NewWriterSize(smtplog, 4096)
}
convo = smtpd.NewConvo(conn, l)
```

这会将具体类型为 *sLogger 的nil传递给期望使用 io.Writer 的对象，从而导致接口转换并隐藏 nil。为了解决这个问题，我们可以添加一个带有 io.Writer 变量的间接级别，该变量必须明确设置：

```golang
var l2 io.Writer

if smtplog != nil {
    l := &sLogger
    l.prefix = logpref
    l.writer = ....
    l2 = l
}
convo = smtpd.NewConvo(conn, l2)

```

如果我们不初始化特殊的日志记录，则 l2 会保持纯净的 io.Writer nil 值，并会在 smtpd 软件包中的代码深度处如此检测。

（您可以通过将设置拉入返回类型为 io.Writer 的函数中，并在不进行日志记录的情况下显式返回nil的方式来做类似的技巧。但是，如果您提供设置，则必须返回接口类型。函数的返回类型为 *sLogger，您将再次遇到相同的问题。）

如果您想在 sLogger 方法函数中保留零防护代码，这是一个好习惯。最后，我决定不想。如果将来我在此代码中遇到这种初始化错误，我希望崩溃，以便对其进行修复。

我从中学到的另一个教训是，如果出于调试目的而打印值时可能会遇到此问题，则我不想使用 %v 作为格式说明符，而是要使用 %#v。前者将同时为 nil 接口和 nil 具体类型的情况打印一个普通且具有误导性的 nil，而后者将为前者打印 nil，并且类似 (*main.fake)(nil) 对于后者。

## 边注栏: 测试程序

```golang
package main

import (
    "fmt"
    "io"
)

type fake struct {
    io.Writer
}

func fred(logger io.Writer) {
    if logger != nil {
        logger.Write([]byte("a test\n"))
    }
}

func main() {
    // this is born <nil>
    var t *fake

    fred(nil)
    fmt.Printf("passed 1\n")
    fred(t)
    fmt.Printf("passed 2\n")
}
```

---

via: https://utcc.utoronto.ca/~cks/space/blog/programming/GoNilNotNil

作者：[ChrisSiebenmann](https://utcc.utoronto.ca/~cks/space/People/ChrisSiebenmann)
译者：[anxk](https://github.com/anxk)
校对：[校对者ID](https://github.com/校对者ID)

本文由 [GCTT](https://github.com/studygolang/GCTT) 原创编译，[Go 中文网](https://studygolang.com/) 荣誉推出
