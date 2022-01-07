package unlimitSizeChan

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRingBuffer(t *testing.T) {
	rb := NewRingBuffer()

	// 空ring读取测试
	v, err := rb.Read()
	assert.Nil(t, v)
	assert.Error(t, err, ErrRingIsEmpty)

	// 写入一个元素，然后读取该元素,做各种判断
	writeValue := 0
	rb.Write(writeValue)
	v, err = rb.Read()
	assert.NoError(t, err)             // 判断err是否为nil
	assert.Equal(t, writeValue, v)     // 判断读取的值是否为写入值
	assert.Equal(t, 1, rb.readCell.r)  // 判断cell的读指针
	assert.Equal(t, 1, rb.writeCell.w) // 判断cell的写指针
	assert.True(t, rb.IsEmpty())       // ring是否空

	var writeSlice []int
	var readSlice []int

	// 写入元素，确保没有扩容
	for i := 1; i <= CellInitialSize; i++ {
		rb.Write(i)
		writeSlice = append(writeSlice, i)
	}
	assert.Equal(t, CellInitialSize*CellInitialCount, rb.Capacity())

	// 继续写入元素，确保已经扩容
	for i := CellInitialSize + 1; i <= CellInitialSize*CellInitialCount; i++ {
		rb.Write(i)
		writeSlice = append(writeSlice, i)
	}
	assert.Equal(t, CellInitialSize*(CellInitialCount+1), rb.Capacity())

	// 读取所有数据
	for {
		v, err := rb.Read()
		if err == ErrRingIsEmpty {
			break
		}
		readSlice = append(readSlice, v.(int))
	}

	// 比较写入元素和读出元素的值和数量和顺序
	assert.Equal(t, len(writeSlice), len(readSlice))
	for i := 0; i < len(writeSlice); i++ {
		assert.Equal(t, writeSlice[i], readSlice[i])
	}

	rb.Reset()
	assert.Equal(t, CellInitialSize*CellInitialCount, rb.Capacity())
	assert.True(t, rb.IsEmpty())

}
