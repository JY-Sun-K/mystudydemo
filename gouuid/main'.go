package main

import (
	"errors"
	"fmt"
	"runtime"
	"sync"
	"time"
)

var sequence int =0
var lastTime int =-1
var workerIdBits =5
var datacenterIdBits =5
var sequenceBits =12

var maxWorkerId int =-1 ^(-1 << workerIdBits)
var maxDatacenterId int =-1 ^(-1<< datacenterIdBits)
var maxSequence int =-1^(-1<<sequenceBits)

var workerIdShift = sequenceBits
var datacenterShift = workerIdBits + sequenceBits
var timestampShift = datacenterIdBits + workerIdBits + sequenceBits

type Snowflake struct {
	datacenterId int
	workerId int
	epoch int
	mt *sync.Mutex
}

func NewSnowflake(datacenterId int,workerId int,epoch int)(*Snowflake,error)  {
	if datacenterId > maxDatacenterId||datacenterId<0 {
		return nil,errors.New(fmt.Sprintf("datacenter cant be greater than %d or less than 0",maxDatacenterId))
	}
	if workerId > maxWorkerId || workerId < 0 {
		return nil, errors.New(fmt.Sprintf("workerId cant be greater than %d or less than 0", maxWorkerId))
	}
	if epoch > getCurrentTime() {
		return nil, errors.New(fmt.Sprintf("epoch time cant be after now"))
	}
	sf:=Snowflake{
		datacenterId: datacenterId,
		workerId:     workerId,
		epoch:        epoch,
		mt:           new(sync.Mutex),
	}
	return &sf,nil

}

func (sf *Snowflake) getUniqueId()int {
	sf.mt.Lock()
	defer sf.mt.Unlock()
	currentTime := getCurrentTime()
	if currentTime<lastTime {
		currentTime =waitUntilNextTime(lastTime)
		sequence=0
	}else if currentTime==lastTime{
		sequence =(sequence+1)&maxSequence
		if sequence == 0 { //overflow max num,wait next time
			currentTime = waitUntilNextTime(lastTime)
		}
	}else if currentTime > lastTime { //next time
		sequence = 0
		lastTime = currentTime
	}
	return (currentTime-sf.epoch)<<timestampShift | sf.datacenterId<<datacenterShift | sf.workerId<<workerIdShift | sequence
}

func waitUntilNextTime(lasttime int) int {
	currentTime := getCurrentTime()
	for currentTime <= lasttime {
		time.Sleep(1 * time.Second / 1000) //sleep micro second
		currentTime = getCurrentTime()
	}
	return currentTime
}

func getCurrentTime() int {
	return int(time.Now().UnixNano() / 1e6) //micro second
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	datacenterId := 0
	workerId := 0
	epoch := 1596850974657
	s, err := NewSnowflake(datacenterId, workerId, epoch)
	if err != nil {
		panic(err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 1000000; i++ {
		wg.Add(1)
		go func() {
			fmt.Println(s.getUniqueId())
			wg.Done()
		}()
	}
	wg.Wait()
}

