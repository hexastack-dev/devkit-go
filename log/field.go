package log

// LogField is an additional structured field
type LogField struct {
	Key   string
	Value any
}

func Field(k string, v any) LogField {
	return LogField{
		Key:   k,
		Value: v,
	}
}
