struct SumReply {
	1: i64 value
}

struct ConcatReply {
	1: string value
}

service AddService {
	SumReply Sum(1: i64 a, 2: i64 b)
	ConcatReply Concat(1: string a, 2: string b)
}
