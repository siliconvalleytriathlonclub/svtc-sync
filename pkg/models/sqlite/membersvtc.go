package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"svtc-sync/pkg/models"

	"github.com/mattn/go-sqlite3"
)

type MemberModel struct {
	DB *sql.DB
}

// --------------------------------------------------------------------------------------------

// Function to add a member record to the database, returns errors on failure to process query, unique field constrain violations
// or other db query failures. Particulars on data formats are
//   - removal of apostrophes (') in names e.g. "O'Connor"
//   - date fileds (joined, expired) are expected to be "YYYY-=MM-DD"
//   - an active flag is used to indicate invalid records (set to false / "0")
func (m *MemberModel) Insert(member *models.MemberSVTC) error {

	// Convert bool to int for sqlite3
	var flag int64
	if member.Active {
		flag = 1
	}

	query := "INSERT INTO member "
	query += "(num, active, login, firstname, middle, lastname, email, status, joined, expired, address, addr_ext, phone, mobile, city, state, zip) "
	query += "VALUES ("
	query += fmt.Sprintf("'%s', ", member.Num)
	query += fmt.Sprintf("%d, ", flag)
	query += fmt.Sprintf("'%s', ", member.Login)
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
			return fmt.Errorf("insert member failed: %w", errors.New("duplicate member"))
		} else {
			return fmt.Errorf("insert member failed: %w", err)
		}
	}

	_, err = result.LastInsertId()
	if err != nil {
		return fmt.Errorf("could not get last inserted id: %w", err)
	}

	return nil
}

// --------------------------------------------------------------------------------------------

// Function to query and return a list of valid members. Based on the specified expire date string
// members will be filtered by status and expire date.
func (m *MemberModel) List(expire string) ([]*models.MemberSVTC, error) {

	query := "SELECT num, firstname, lastname, email, status, expired "
	query += "FROM member "
	query += "WHERE active = ? "

	if expire != "1963-11-04" {
		query += "AND status = ? "
		query += "AND expired > ? "
	}

	// fmt.Printf("%s \n", query)

	rows, err := m.DB.Query(query, 1, "Expired", expire)
	if err != nil {
		return nil, fmt.Errorf("sql query failed: %w", err)
	}
	defer rows.Close()

	// memberList := make([]MemberSVTC, 0)
	memberList := []*models.MemberSVTC{}

	for rows.Next() {
		// Create a pointer to a new struct.
		member := &models.MemberSVTC{}

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

// Function to retrieve a single member record based on their member number
func (m *MemberModel) Get(num string) (*models.MemberSVTC, error) {

	member := &models.MemberSVTC{}

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
			return nil, err
		} else {
			return nil, fmt.Errorf("member sql query failed: %w", err)
		}
	}

	return member, nil

}

// --------------------------------------------------------------------------------------------

// Function to update a member's status based on their member number
func (m *MemberModel) UpdateStatus(num, status, expired string) error {

	query := "UPDATE member SET status = ?, expired = ? WHERE num = ?"

	stmt, err := m.DB.Prepare(query)
	if err != nil {
		return fmt.Errorf("prepare sql query failed: %w", err)
	}
	defer stmt.Close()

	result, err := stmt.Exec(status, expired, num)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("sql query failed for %s: %w", num, errors.New("no matching record found"))
		} else {
			return fmt.Errorf("sql query failed for %s: %w", num, err)
		}
	}

	// sql.ErrNoRows does not work here. Evaluate result.RowsAffected() instead.
	_, err = result.RowsAffected()
	if err != nil {
		return fmt.Errorf("could not get rows affected: %w", err)
	}

	return nil
}

// --------------------------------------------------------------------------------------------
