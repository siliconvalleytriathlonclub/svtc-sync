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

// Function to add a member record to the database, returns errors on failure to process query, unique field constraint violations
// or other db query failures. Particulars on data formats are
//   - escape apostrophes (') in names e.g. "O'Connor"
//   - date fileds (joined, expired) are expected to be "YYYY-MM-DD"
//   - an active flag is used to indicate invalid records (set to false / "0")
func (m *MemberModel) Insert(member *models.MemberSVTC) error {

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

	stmt, err := m.DB.Prepare(query)
	if err != nil {
		return fmt.Errorf("prepare sql query failed: %w", err)
	}
	defer stmt.Close()

	result, err := stmt.Exec()
	if err != nil {
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

// Function to query and return a list of valid members filtered by the provided search member struct.
// Based on the specified expire date string members will be filtered by status and expire date.
func (m *MemberModel) ListMembers() ([]*models.MemberSVTC, error) {

	query := "SELECT num, firstname, lastname, email, status, expired "
	query += "FROM member "
	query += "WHERE active = ? "

	rows, err := m.DB.Query(query, 1)
	if err != nil {
		return nil, fmt.Errorf("sql query failed: %w", err)
	}
	defer rows.Close()

	memberList := []*models.MemberSVTC{}

	for rows.Next() {

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

// Function to query and return a list of valid members filtered by the provided search member struct.
// Based on the specified expire date string members will be filtered by status and expire date.
func (m *MemberModel) ListMatch(platform string, search *models.MemberSVTC) ([]*models.MemberSVTC, error) {

	var query, lnamestr string

	switch platform {

	case "strava":

		// The query for Strava uses Firstname and Initial of Lastname
		// Example:
		// 		select *
		//		from member
		// 		where active = 1
		// 		and (firstname = 'Dave' and lastname like 'S%')
		//		and status = 'Expired'
		// 		and expired > '2001-01-31';

		query = "SELECT num, firstname, lastname, email, status, expired "
		query += "FROM member "
		query += "WHERE active = ? "
		query += "AND ((lower(firstname) = ? AND lower(lastname) LIKE ?) OR lower(email) = ?) "

		if search.Status != "" {
			query += "AND status = ? "
		}

		if search.Expired != "1963-11-04" {
			query += "AND expired > ? "
		}

		lnamestr = search.LastName + string('%')

	case "slack":

		// The query for Slack uses (Firstname and Lastname) OR Email
		// Example:
		// 		select *
		//		from member
		// 		where active = 1
		// 		and (firstname = 'Dave' and lastname = 'Scott') or (email = 'theman@gmail.com')
		//		and status = 'Expired'
		// 		and expired > '2001-01-31';

		query = "SELECT num, firstname, lastname, email, status, expired "
		query += "FROM member "
		query += "WHERE active = ? "
		query += "AND ((lower(firstname) = ? AND lower(lastname) = ?) OR lower(email) = ?) "

		if search.Status != "" {
			query += "AND status = ? "
		}

		if search.Expired != "1963-11-04" {
			query += "AND expired > ? "
		}

		lnamestr = search.LastName

	}

	// Note: Expired status and expire date remain ignored if not added to query string above
	rows, err := m.DB.Query(query, 1, search.FirstName, lnamestr, search.Email, search.Status, search.Expired)
	if err != nil {
		return nil, fmt.Errorf("sql query failed: %w", err)
	}
	defer rows.Close()

	memberList := []*models.MemberSVTC{}

	for rows.Next() {

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

// Function to query and return a list of valid members filtered by the provided search member struct.
// Based on the specified expire date string members will be filtered by status and expire date.
func (m *MemberModel) ListAlias() ([]*models.MemberSVTC, []*models.MemberAlias, error) {

	query := "SELECT member.num, member.firstname as mf, member.lastname as ml, member.email as me, "
	query += "alias.firstname as af, alias.lastname as al, alias.email as ae "
	query += "FROM member INNER JOIN alias ON member.id = alias.memberid "

	rows, err := m.DB.Query(query)
	if err != nil {
		return nil, nil, fmt.Errorf("sql query failed: %w", err)
	}
	defer rows.Close()

	memberList := []*models.MemberSVTC{}
	aliasList := []*models.MemberAlias{}

	for rows.Next() {

		member := &models.MemberSVTC{}
		alias := &models.MemberAlias{}

		err = rows.Scan(
			&member.Num,
			&member.FirstName,
			&member.LastName,
			&member.Email,
			&alias.FirstName,
			&alias.LastName,
			&alias.Email,
		)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, nil, fmt.Errorf("member sql query failed: %w", errors.New("no matching record found"))
			} else {
				return nil, nil, fmt.Errorf("member sql query failed: %w", err)
			}
		}

		memberList = append(memberList, member)
		aliasList = append(aliasList, alias)

	}

	err = rows.Err()
	if err != nil {
		return nil, nil, fmt.Errorf("row iteraton error: %w", err)
	}

	return memberList, aliasList, nil

}

// --------------------------------------------------------------------------------------------

// Function to query and return a list of valid members filtered by the provided search member struct.
// Based on the specified expire date string members will be filtered by status and expire date.
func (m *MemberModel) GetAlias(search *models.MemberSVTC) ([]*models.MemberSVTC, error) {

	// Example:
	// 		select *
	// 		from member
	// 		inner join alias on member.id = alias.memberid
	// 		where member.active = 1
	// 		and (alias.firstname = 'Dave' and alias.lastname = 'Scott') or (alias.email = 'theman@gmail.com')
	//		and member.status = 'Expired'
	// 		and member.expired > '2001-01-31';

	query := "SELECT member.num, member.firstname, member.lastname, member.email, member.status, member.expired "
	query += "FROM member INNER JOIN alias ON member.id = alias.memberid "
	query += "WHERE member.active = ? "
	query += "AND ((lower(alias.firstname) = ? AND lower(alias.lastname) = ?) OR lower(alias.email) = ?) "

	if search.Status != "" {
		query += "AND member.status = ? "
	}

	if search.Expired != "1963-11-04" {
		query += "AND member.expired > ? "
	}

	rows, err := m.DB.Query(query, 1, search.FirstName, search.LastName, search.Email, search.Status, search.Expired)
	if err != nil {
		return nil, fmt.Errorf("sql query failed: %w", err)
	}
	defer rows.Close()

	memberList := []*models.MemberSVTC{}

	for rows.Next() {

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

	_, err = result.RowsAffected()
	if err != nil {
		return fmt.Errorf("could not get rows affected: %w", err)
	}

	return nil
}

// --------------------------------------------------------------------------------------------
