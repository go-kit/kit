struct AddReply {
	1: i64 value
}

service AddService {
	AddReply Add(1: i64 a, 2: i64 b)
}
