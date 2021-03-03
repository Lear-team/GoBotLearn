package sqlapi

import (
	"context"
	"strings"
	"time"

	apitypes "GoBotPigeon/types/apitypes"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

// CheckingPigeonWork ...
func (api *API) CheckingPigeonWork(ctx context.Context, userID string) (bool, error) {

	var botWork = []apitypes.BotWork{}

	query := "SELECT * FROM prj_botwork WHERE userid = $1 LIMIT 1;"

	err := api.db.SelectContext(ctx, &botWork, query, userID)
	if err != nil {
		return false, errors.Wrap(err, "SELECT * FROM prj_botwork failed")
	}
	if len(botWork) == 0 {
		return false, err
	}

	return botWork[0].BotWorkFlag, err
}

// StartPigeonWork ...
func (api *API) StartPigeonWork(ctx context.Context, userID string) error { // проверить как работает

	err := api.CreatePigeonWorkFlag(ctx, userID)
	if err != nil {
		return errors.Wrap(err, "Create pigeon work flag failed")
	}

	return nil
}

// StopPigeonWork ...
func (api *API) StopPigeonWork(ctx context.Context, userID string) error {
	work := func(ctx context.Context, db TxContext) error {

		query := `INSERT INTO prj_botwork ("botworkid", "userid", "botworkflag") VALUES ($1, $2, $3)
					ON CONFLICT (userid)
					DO 
					UPDATE SET botworkflag = $3`

		uuidWithHyphen := uuid.New()
		uid := strings.Replace(uuidWithHyphen.String(), "-", "", -1)

		if _, err := db.ExecContext(ctx, query, uid, userID, false); err != nil {
			return err
		}
		return nil
	}

	if err := RunInTransaction(ctx, api.db, work); err != nil {
		return err
	}

	return nil
}

// CreatePigeonWorkFlag ...
func (api *API) CreatePigeonWorkFlag(ctx context.Context, userID string) error {
	work := func(ctx context.Context, db TxContext) error {

		query := `INSERT INTO prj_botwork ("botworkid", "userid", "botworkflag") VALUES ($1, $2, $3)
					ON CONFLICT (userid)
					DO 
					UPDATE SET botworkflag = $3`

		uuidWithHyphen := uuid.New()
		uid := strings.Replace(uuidWithHyphen.String(), "-", "", -1)

		if _, err := db.ExecContext(ctx, query, uid, userID, true); err != nil {
			return err
		}
		return nil
	}

	if err := RunInTransaction(ctx, api.db, work); err != nil {
		return err
	}

	return nil
}

// SetLastComandUser ...
func (api *API) SetLastComandUser(ctx context.Context, userN string, command string) error {
	user, err := api.GetUserByName(ctx, userN)
	if err != nil {
		return errors.Wrap(err, "Get user by name failed")
	}

	work := func(ctx context.Context, db TxContext) error {

		query := `INSERT INTO prj_lastusercommand ("commandid", "userid", "command", "datacommand") 
					VALUES ($1, $2, $3, $4)`

		uuidWithHyphen := uuid.New()
		uid := strings.Replace(uuidWithHyphen.String(), "-", "", -1)
		today := time.Now()
		tTime := today.Add(10 * time.Minute).Format("2006/1/2 15:04")
		if _, err := db.ExecContext(ctx, query, uid, user.UserID, command, tTime); err != nil {
			return err
		}
		return nil
	}
	if err := RunInTransaction(ctx, api.db, work); err != nil {
		return err
	}
	return nil
}

// GetLastCommandByUserName ...
func (api *API) GetLastCommandByUserName(ctx context.Context, userN string) (*apitypes.LastUserCommand, error) {
	return getLastCommandByUserName(ctx, api.db, userN)
}

func getLastCommandByUserName(ctx context.Context, db TxContext, userN string) (*apitypes.LastUserCommand, error) {
	user, err := getUserByName(ctx, db, userN)

	if err != nil {
		return nil, errors.Wrap(err, "Get user by name failed")
	}

	if user == nil {
		return nil, nil // создать ошибку
	}

	var arrCommand = []apitypes.LastUserCommand{}

	query := `SELECT commandid, userid, command, datacommand FROM prj_lastusercommand 
				WHERE (userid = $1) ORDER BY datacommand DESC`

	err = db.SelectContext(ctx, &arrCommand, query, user.UserID)
	if err != nil {
		return nil, errors.Wrap(err, "SELECT * FROM prj_lastusercommand failed")
	}

	if len(arrCommand) == 0 {
		return nil, nil // создать ошибку
	}

	return &arrCommand[0], nil
}

// DeleteLastCommand ...
func (api *API) DeleteLastCommand(ctx context.Context, userId string, command string) error {

	work := func(ctx context.Context, db TxContext) error {

		query := `DELETE FROM prj_lastusercommand WHERE userid = $1`

		if _, err := db.ExecContext(ctx, query, userId); err != nil {
			return errors.Wrap(err, "DELETE FROM prj_lastusercommand")
		}
		return nil
	}

	if err := RunInTransaction(ctx, api.db, work); err != nil {
		return err
	}
	return nil
}
