package Engine

import "fmt"

type PoolItem[T any] struct {
	ItemIndex int // FOR INTERNAL USE ONLY MUST NEVER BE CHANGED BY THE IMPLEMENTATION
	Item      T
}

type Pool[T any] struct {
	PoolName string

	TotalNumItemsInPool    int
	CurNumAliveItemsInPool int

	Items []*PoolItem[T]
}

func (pool *Pool[T]) InitPool(poolName string, totalNumItemsInPool int) {

	pool.PoolName = poolName
	pool.TotalNumItemsInPool = totalNumItemsInPool
	pool.CurNumAliveItemsInPool = 0

	pool.Items = make([]*PoolItem[T], pool.TotalNumItemsInPool)

	for index := range pool.TotalNumItemsInPool {
		pool.Items[index] = &PoolItem[T]{
			ItemIndex: index,
		}
	}
}

func (pool *Pool[T]) GetAnUnusedItemFromPool() *PoolItem[T] {
	pool.CurNumAliveItemsInPool++

	if pool.CurNumAliveItemsInPool-1 >= pool.TotalNumItemsInPool {
		fmt.Println("Tried to acess more than created items in pool : ", pool.PoolName, ". New Item Index : ", pool.CurNumAliveItemsInPool-1, ". Total num items created at init : ", pool.TotalNumItemsInPool)
	}

	pool.Items[pool.CurNumAliveItemsInPool-1].ItemIndex = pool.CurNumAliveItemsInPool - 1

	return pool.Items[pool.CurNumAliveItemsInPool-1]
}

func (pool *Pool[T]) KillItemInPool(poolItemToKill *PoolItem[T]) {

	if pool.CurNumAliveItemsInPool == 1 || poolItemToKill.ItemIndex == pool.CurNumAliveItemsInPool-1 {
		pool.CurNumAliveItemsInPool--
		return
	}

	lastAlivePoolItemIndex := pool.CurNumAliveItemsInPool - 1
	curPoolItemIndex := poolItemToKill.ItemIndex

	pool.Items[curPoolItemIndex], pool.Items[lastAlivePoolItemIndex] = pool.Items[lastAlivePoolItemIndex], pool.Items[curPoolItemIndex]
	pool.Items[curPoolItemIndex].ItemIndex = curPoolItemIndex

	pool.CurNumAliveItemsInPool--
}

func (pool *Pool[T]) KillAllItemsInPool() {
	pool.CurNumAliveItemsInPool = 0
}

type OperationOnPoolItem[T any] func(curPoolItem *PoolItem[T])

func (pool *Pool[T]) PerformOperationOnAlivePoolItems(operationOnPoolItem OperationOnPoolItem[T]) {
	for index := range pool.CurNumAliveItemsInPool {
		operationOnPoolItem(pool.Items[index])
	}
}

func (pool *Pool[T]) PerformOperationOnAlivePoolItemsBackwards(operationOnPoolItem OperationOnPoolItem[T]) {

	for index := pool.CurNumAliveItemsInPool - 1; index >= 0; index-- {
		operationOnPoolItem(pool.Items[index])
	}
}
