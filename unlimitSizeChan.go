package unlimitSizeChan

import "sync/atomic"

type T interface{}

// UnlimitSizeChan 无限缓存的Channle
type UnlimitSizeChan struct {
	bufCount int64       // 统计元素数量，原子操作
	In       chan<- T    // 写入channle
	Out      <-chan T    // 读取channle
	buffer   *RingBuffer // 自适应扩缩容Buf
}

// Len uc中总共的元素数量
func (uc UnlimitSizeChan) Len() int {
	return len(uc.In) + uc.BufLen() + len(uc.Out)
}

// BufLen uc的buf中的元素数量
func (uc UnlimitSizeChan) BufLen() int {
	return int(atomic.LoadInt64(&uc.bufCount))
}

// NewUnlimitSizeChan 新建一个无限缓存的Channle，并指定In和Out大小(In和Out设置得一样大)
func NewUnlimitSizeChan(initCapacity int) *UnlimitSizeChan {
	return NewUnlitSizeChanSize(initCapacity, initCapacity)
}

// NewUnlitSizeChanSize 新建一个无限缓存的Channle，并指定In和Out大小(In和Out设置得不一样大)
func NewUnlitSizeChanSize(initInCapacity, initOutCapacity int) *UnlimitSizeChan {
	in := make(chan T, initInCapacity)
	out := make(chan T, initOutCapacity)
	ch := UnlimitSizeChan{In: in, Out: out, buffer: NewRingBuffer()}

	go process(in, out, &ch)

	return &ch
}

// 内部Worker Groutine实现
func process(in, out chan T, ch *UnlimitSizeChan) {
	defer close(out) // in 关闭，数据读取后也把out关闭

	// 不断从in中读取数据放入到out或者ringbuf中
loop:
	for {
		// 第一步：从in中读取数据
		value, ok := <-in
		if !ok {
			// in 关闭了，退出loop
			break loop
		}

		// 第二步：将数据存储到out或者buf中
		if atomic.LoadInt64(&ch.bufCount) > 0 {
			// 当buf中有数据时，新数据优先存放到buf中，确保数据FIFO原则
			ch.buffer.Write(value)
			atomic.AddInt64(&ch.bufCount, 1)
		} else {
			// out 没有满,数据放入out中
			select {
			case out <- value:
				continue
			default:
			}

			// out 满了，数据放入buf中
			ch.buffer.Write(value)
			atomic.AddInt64(&ch.bufCount, 1)
		}

		// 第三步：处理buf，一直尝试把buf中的数据放入到out中，直到buf中没有数据
		for !ch.buffer.IsEmpty() {
			select {
			// 为了避免阻塞in，还要尝试从in中读取数据
			case val, ok := <-in:
				if !ok {
					// in 关闭了，退出loop
					break loop
				}
				// 因为这个时候out是满的，新数据直接放入buf中
				ch.buffer.Write(val)
				atomic.AddInt64(&ch.bufCount, 1)

			// 将buf中数据放入out
			case out <- ch.buffer.Peek():
				ch.buffer.Pop()
				atomic.AddInt64(&ch.bufCount, -1)

				if ch.buffer.IsEmpty() { // 避免内存泄露
					ch.buffer.Reset()
					atomic.StoreInt64(&ch.bufCount, 0)
				}
			}
		}
	}

	// in被关闭退出loop后，buf中还有可能有未处理的数据，将他们塞入out中，并重置buf
	for !ch.buffer.IsEmpty() {
		out <- ch.buffer.Pop()
		atomic.AddInt64(&ch.bufCount, -1)
	}
	ch.buffer.Reset()
	atomic.StoreInt64(&ch.bufCount, 0)
}
