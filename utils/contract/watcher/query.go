package watcher

func QueryFromBytes(bytes [][]byte) []interface{} {
	var queries []interface{}
	for _, _byte := range bytes {
		queries = append(queries, _byte)
	}
	return queries
}

func QueryFromBytes32(bytes [][32]byte) []interface{} {
	var queries []interface{}
	for _, _byte := range bytes {
		queries = append(queries, _byte)
	}
	return queries
}
