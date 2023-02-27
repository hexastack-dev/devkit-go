package main

import (
	"context"
	"fmt"
	"log"

	"github.com/hexastack-dev/devkit-go/data/optional"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

type Contact struct {
	Id        int64
	FirstName string                 `db:"first_name"`
	LastName  optional.Value[string] `db:"last_name"`
	Addresses []Address
}

type Address struct {
	Id         int64
	Name       string
	Street     string
	PostalCode optional.Value[int64] `db:"postal_code"`
	ContactId  int64                 `db:"contact_id"`
}

// describe relationship between contact with address,
// which is one to many with left join (possibly null).
type contactAddresses struct {
	Contact `db:"contact"`
	Address struct {
		Id         optional.Value[int64]
		Name       optional.Value[string]
		Street     optional.Value[string]
		PostalCode optional.Value[int64] `db:"postal_code"`
		ContactId  optional.Value[int64] `db:"contact_id"`
	} `db:"address"`
}

const getContacts = `--
SELECT c.id AS "contact.id", c.first_name AS "contact.first_name", c.last_name AS "contact.last_name",
ad.id AS "address.id", ad.name AS "address.name", ad.street AS "address.street", ad.postal_code AS "address.postal_code", ad.contact_id AS "address.contact_id"
FROM contacts c
LEFT JOIN addresses ad ON c.id = ad.contact_id`

const namedCreateContact = `--
INSERT INTO contacts(first_name, last_name)
VALUES (:first_name, :last_name)`

func main() {
	db, err := sqlx.Connect("pgx", "postgres://labs:labs@localhost:5432/labs")
	if err != nil {
		log.Fatalln(err)
	}
	// contact := Contact{
	// 	FirstName: "Melo",
	// 	LastName:  optional.Nil[string](),
	// }
	// _, err = db.NamedExec(namedCreateContact, contact)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	contacts, err := findAll(db)
	if err != nil {
		log.Fatal(err)
	}
	for _, contact := range contacts {
		log.Println("Contact:")
		log.Printf("id:%d\tfirstName:%s\tlastName:%s\n", contact.Id, contact.FirstName, OptString(contact.LastName))
		if contact.Addresses == nil {
			log.Println("Addresses is nil")
		} else {
			log.Println("Addresses")
			for _, addr := range contact.Addresses {
				log.Printf("id:%d\tname:%s\tstreet:%s\tpostalCode:%s\tcontactId:%d\n", addr.Id, addr.Name, addr.Street, OptString(addr.PostalCode), addr.ContactId)
			}
		}
	}
}

func findAll(db *sqlx.DB) ([]Contact, error) {
	var rows []contactAddresses
	if err := db.SelectContext(context.Background(), &rows, getContacts); err != nil {
		return nil, err
	}

	contactMap := make(map[int64]Contact)
	for _, row := range rows {
		contact, ok := contactMap[row.Contact.Id]
		if !ok {
			contact = row.Contact
		}
		if row.Address.Id.IsNotNil() {
			address := row.Address
			contact.Addresses = append(contactMap[row.Contact.Id].Addresses, Address{
				Id:         address.Id.Val(),
				Name:       address.Name.Val(),
				Street:     address.Street.Val(),
				PostalCode: address.PostalCode,
				ContactId:  address.ContactId.Val(),
			})
		}

		contactMap[row.Contact.Id] = contact
	}

	contacts := make([]Contact, 0, len(contactMap))
	for _, contact := range contactMap {
		contacts = append(contacts, contact)
	}

	return contacts, nil
}

func OptString[T comparable](opt optional.Value[T]) string {
	if opt.IsNil() {
		return "null"
	}
	return fmt.Sprint(opt.Val())
}
