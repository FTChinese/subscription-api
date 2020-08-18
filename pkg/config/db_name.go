package config

type SubsDB int

const (
	SubsDBSandbox SubsDB = iota
	SubsDBProd
)

var subsDBNames = [...]string{
	"sandbox",
	"premium",
}

func (x SubsDB) String() string {
	if x < SubsDBSandbox || x > SubsDBProd {
		return subsDBNames[1]
	}

	return subsDBNames[x]
}
