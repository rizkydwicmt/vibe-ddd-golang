package database

/*----------- DirectionEnum -----------*/

type DirectionEnum string

const (
	ASC  DirectionEnum = "asc"
	DESC DirectionEnum = "desc"
)

func (e DirectionEnum) ToString() string {
	switch e {
	case ASC:
		return "asc"
	case DESC:
		return "desc"
	}
	return ""
}

func (e DirectionEnum) IsValid() bool {
	switch e {
	case ASC, DESC:
		return true
	}
	return false
}

/*----------- DriverEnum -----------*/

type DriverEnum string

const (
	POSTGRES DriverEnum = "postgres"
	MYSQL    DriverEnum = "mysql"
)

func (e DriverEnum) ToString() string {
	switch e {
	case POSTGRES:
		return "postgres"
	case MYSQL:
		return "mysql"
	default:
		return ""
	}
}

func (e DriverEnum) ToShortString() string {
	switch e {
	case POSTGRES:
		return "pg"
	case MYSQL:
		return "sql"
	default:
		return ""
	}
}

func (e DriverEnum) IsValid() bool {
	switch e {
	case POSTGRES, MYSQL:
		return true
	}
	return false
}
