package postoffice

// Parcel contains the data to compose an email
type Parcel struct {
	FromAddress string
	FromName    string
	ToAddress   string
	ToName      string
	Subject     string
	Body        string
}
