# unlimitSizeChan

[![License](https://img.shields.io/:license-MIT-blue.svg)](https://opensource.org/licenses/MIT)

无限缓冲的channel，golang实现

请参阅以下文章和问题：
1. https://github.com/golang/go/issues/20352
2. https://stackoverflow.com/questions/41906146/why-go-channels-limit-the-buffer-size
3. https://medium.com/capital-one-tech/building-an-unbounded-channel-in-go-789e175cd2cd
4. https://erikwinter.nl/articles/2020/channel-with-infinite-buffer-in-golang/


## Usage
```go
ch := NewUnlimitSizeChan(1000)
// or ch := NewUnlitSizeChanSize(100,200)

go func() {
    for ...... {
        ...
        ch.In <- ... // send values
        ...
    }

    close(ch.In) // close In channel
}()


for v := range ch.Out { // read values
    fmt.Println(v)
}
```


> 设计和实现思路请参考作者博客: https://www.cnblogs.com/yinbiao/p/15784545.html


