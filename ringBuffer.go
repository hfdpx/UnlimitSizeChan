package unlimitSizeChan

import (
	"errors"
)

var ErrRingIsEmpty = errors.New("ringbuffer is empty")

// CellInitialSize cell的初始容量
var CellInitialSize = 1024

// CellInitialCount 初始化cell数量
var CellInitialCount = 2

type cell struct {
	Data     []T   // 数据部分
	fullFlag bool  // cell满的标志
	next     *cell // 指向后一个cellBuffer
	pre      *cell // 指向前一个cellBuffer

	r int // 下一个要读的指针
	w int // 下一个要下的指针
}

type RingBuffer struct {
	cellCount int // cell 数量统计

	readCell  *cell // 下一个要读的cell
	writeCell *cell // 下一个要写的cell
}

// NewRingBuffer 新建一个ringbuffe，包含两个cell
func NewRingBuffer() *RingBuffer {
	rootCell := &cell{
		Data: make([]T, CellInitialSize),
	}
	lastCell := &cell{
		Data: make([]T, CellInitialSize),
	}
	rootCell.pre = lastCell
	lastCell.pre = rootCell
	rootCell.next = lastCell
	lastCell.next = rootCell

	return &RingBuffer{
		cellCount: CellInitialCount,
		readCell:  rootCell,
		writeCell: rootCell,
	}
}

// Read 读取数据
func (r *RingBuffer) Read() (T, error) {
	// 无数据
	if r.IsEmpty() {
		return nil, ErrRingIsEmpty
	}

	// 读取数据，并将读指针向右移动一位
	value := r.readCell.Data[r.readCell.r]
	r.readCell.r++

	// 此cell已经读完
	if r.readCell.r == CellInitialSize {
		// 读指针归零，并将该cell状态置为非满
		r.readCell.r = 0
		r.readCell.fullFlag = false
		// 将readCell指向下一个cell
		r.readCell = r.readCell.next

	}

	return value, nil
}

// Pop 读一个元素，读完后移动指针
func (r *RingBuffer) Pop() T {
	value, err := r.Read()
	if err != nil {
		panic(err.Error())
	}
	return value
}

// Peek 窥视 读一个元素，仅读但不移动指针
func (r *RingBuffer) Peek() T {
	if r.IsEmpty() {
		panic(ErrRingIsEmpty.Error())
	}

	// 仅读
	value := r.readCell.Data[r.readCell.r]
	return value
}

// Write 写入数据
func (r *RingBuffer) Write(value T) {
	// 在 r.writeCell.w 位置写入数据，指针向右移动一位
	r.writeCell.Data[r.writeCell.w] = value
	r.writeCell.w++

	// 当前cell写满了
	if r.writeCell.w == CellInitialSize {
		// 指针置0，将该cell标记为已满，并指向下一个cell
		r.writeCell.w = 0
		r.writeCell.fullFlag = true
		r.writeCell = r.writeCell.next
	}

	// 下一个cell也已满，扩容
	if r.writeCell.fullFlag == true {
		r.grow()
	}

}

// grow 扩容
func (r *RingBuffer) grow() {
	// 新建一个cell
	newCell := &cell{
		Data: make([]T, CellInitialSize),
	}

	// 总共三个cell，writeCell，preCell，newCell
	// 本来关系： preCell <===> writeCell
	// 现在将newcell插入：preCell <===> newCell <===> writeCell
	pre := r.writeCell.pre
	pre.next = newCell
	newCell.pre = pre
	newCell.next = r.writeCell
	r.writeCell.pre = newCell

	// 将writeCell指向新建的cell
	r.writeCell = r.writeCell.pre

	// cell 数量加一
	r.cellCount++
}

// IsEmpty 判断ringbuffer是否为空
func (r *RingBuffer) IsEmpty() bool {
	// readCell和writeCell指向同一个cell，并且该cell的读写指针也指向同一个位置，并且cell状态为非满
	if r.readCell == r.writeCell && r.readCell.r == r.readCell.w && r.readCell.fullFlag == false {
		return true
	}
	return false
}

// Capacity ringBuffer容量
func (r *RingBuffer) Capacity() int {
	return r.cellCount * CellInitialSize
}

// Reset 重置为仅指向两个cell的ring
func (r *RingBuffer) Reset() {

	lastCell := r.readCell.next

	lastCell.w = 0
	lastCell.r = 0
	r.readCell.r = 0
	r.readCell.w = 0
	r.cellCount = CellInitialCount

	lastCell.next = r.readCell
}
