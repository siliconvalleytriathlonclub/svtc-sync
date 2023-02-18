package main

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/mattn/go-sqlite3"
)

type ClubSQLModel struct {
	DB *sql.DB
}

// --------------------------------------------------------------------------------------------
//
// Test run to insert all records read from refernce CSV into local Sqlite3 DB
/*
	for _, m := range mlCSV {
		err = app.clubSQL.Insert(m)
		if err != nil {
			log.Printf("[Insert] %s", err)
		}
	}
	log.Printf("[main] TEST Insert all records read from CSV into local Sqlite3 DB")
*/
//

func (m *ClubSQLModel) Insert(member *MemberSVTC) error {

	// Convert bool to int for sqlite3
	var flag int64
	if member.Active {
		flag = 1
	}

	query := "INSERT INTO member "
	query += "(num, active, firstname, middle, lastname, email, status, joined, expired, address, addr_ext, phone, mobile, city, state, zip) "
	query += "VALUES ("
	query += fmt.Sprintf("%d, ", member.Num)
	query += fmt.Sprintf("%d, ", flag)
	query += fmt.Sprintf("'%s', ", member.FirstName)
	query += fmt.Sprintf("'%s', ", member.Middle)
	query += fmt.Sprintf("'%s', ", strings.ReplaceAll(member.LastName, "'", "''"))
	query += fmt.Sprintf("'%s', ", member.Email)
	query += fmt.Sprintf("'%s', ", member.Status)
	query += fmt.Sprintf("'%s', ", member.Joined)
	query += fmt.Sprintf("'%s', ", member.Expired)
	query += fmt.Sprintf("'%s', ", strings.ReplaceAll(member.Address, "'", "''"))
	query += fmt.Sprintf("'%s', ", member.AddrExt)
	query += fmt.Sprintf("'%s', ", member.Phone)
	query += fmt.Sprintf("'%s', ", member.Mobile)
	query += fmt.Sprintf("'%s', ", member.City)
	query += fmt.Sprintf("'%s', ", member.State)
	query += fmt.Sprintf("'%s'", member.Zip)
	query += ")"

	// fmt.Printf("%s \n", query)

	stmt, err := m.DB.Prepare(query)
	if err != nil {
		return fmt.Errorf("prepare sql query failed: %w", err)
	}
	defer stmt.Close()

	result, err := stmt.Exec()
	if err != nil {
		// For reference https://github.com/mattn/go-sqlite3/blob/master/error.go
		sqliteErr := err.(sqlite3.Error)
		if sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return fmt.Errorf("insert goal failed: %w", errors.New("duplicate member"))
		} else {
			return fmt.Errorf("insert goal failed: %w", err)
		}
	}

	_, err = result.LastInsertId()
	if err != nil {
		return fmt.Errorf("could not get last inserted id: %w", err)
	}
	// log.Printf("[insert] Added member [%d] with ID: %d \n", member.Num, lastid)

	// fmt.Printf("%s \n", query)

	return nil
}

// --------------------------------------------------------------------------------------------
//
// Get list of members from Sqlite3 DB
/*
	mlSQL, err := app.clubSQL.MemberList()
	if err != nil {
		log.Printf("[ListMembers] %s", err)
		return
	}
	log.Printf("[main] Read list of %d club members from %s", len(mlSQL), cfg.dbfile)
*/
//

func (m *ClubSQLModel) MemberList() ([]*MemberSVTC, error) {

	query := "SELECT num, firstname, lastname, email, status, expired "
	query += "FROM member "
	query += "WHERE lastname = ?"

	// fmt.Printf("%s \n", query)

	rows, err := m.DB.Query(query, "Apitz")
	if err != nil {
		return nil, fmt.Errorf("sql query failed: %w", err)
	}
	defer rows.Close()

	// memberList := make([]MemberSVTC, 0)
	memberList := []*MemberSVTC{}

	for rows.Next() {
		// Create a pointer to a new struct.
		member := &MemberSVTC{}

		err = rows.Scan(
			&member.Num,
			&member.FirstName,
			&member.LastName,
			&member.Email,
			&member.Status,
			&member.Expired,
		)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, fmt.Errorf("member sql query failed: %w", errors.New("no matching record found"))
			} else {
				return nil, fmt.Errorf("member sql query failed: %w", err)
			}
		}

		memberList = append(memberList, member)

	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("row iteraton error: %w", err)
	}

	return memberList, nil

}

// --------------------------------------------------------------------------------------------
//
/*
	mSQL, err := app.clubSQL.Get(m.Num)
	if err != nil {
		log.Printf("[GetMember] %s", err)
		return
	}
	fmt.Printf("\tSQL: [%d] %s %s (%s) - %s [%s] \n", mSQL.Num, mSQL.FirstName, mSQL.LastName, mSQL.Email, mSQL.Status, mSQL.Expired)
*/
//
func (m *ClubSQLModel) Get(num int) (*MemberSVTC, error) {

	member := &MemberSVTC{}

	query := "SELECT num, firstname, lastname, email, status, expired "
	query += "FROM member "
	query += "WHERE num = ?"

	err := m.DB.QueryRow(query, num).Scan(
		&member.Num,
		&member.FirstName,
		&member.LastName,
		&member.Email,
		&member.Status,
		&member.Expired,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("member sql query failed: %w", errors.New("no matching record found"))
		} else {
			return nil, fmt.Errorf("member sql query failed: %w", err)
		}
	}

	return member, nil

}

// --------------------------------------------------------------------------------------------
