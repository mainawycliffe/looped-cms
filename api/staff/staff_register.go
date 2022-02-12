package staff

import (
	"context"
	"fmt"
	"time"

	"github.com/looped-dev/cms/api/emails"
	"github.com/looped-dev/cms/api/graph/model"
	"github.com/looped-dev/cms/api/models"
	"github.com/looped-dev/cms/api/utils"
	mail "github.com/xhit/go-simple-mail/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Staff struct {
	SMTPClient *mail.SMTPClient
	DBClient   *mongo.Client
}

func NewStaff(smtpClient *mail.SMTPClient, dbClient *mongo.Client) *Staff {
	return &Staff{
		SMTPClient: smtpClient,
		DBClient:   dbClient,
	}
}

// StaffRegister creates a new staff (admin users) and returns the Staff object.
func StaffRegister(client *mongo.Client, input *model.StaffRegisterInput) (*models.Staff, error) {
	hashedPassword, err := utils.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}
	createdAt := primitive.Timestamp{
		T: uint32(time.Now().Unix()),
	}
	staff := &models.Staff{
		Name:           input.Name,
		Email:          input.Email,
		HashedPassword: hashedPassword,
		EmailVerified:  false,
		CreatedAt:      createdAt,
		UpdatedAt:      createdAt,
	}
	result, err := client.Database("cms").Collection("staff").InsertOne(context.TODO(), staff)
	if err != nil {
		return nil, err
	}
	staff.ID = result.InsertedID.(primitive.ObjectID)
	return staff, nil
}

// StaffSendInvite creates a new staff, with a specific role and creates an invite
// code and sends an email to the staff member.
func (s Staff) StaffSendInvite(ctx context.Context, input *model.StaffInviteInput) (*models.Staff, error) {
	code := utils.GenerateInviteCode()
	staff := &models.Staff{
		Email: input.Email,
		Role:  input.Role,
		InviteCode: models.InviteCode{
			Code: code,
			Expiry: primitive.Timestamp{
				T: uint32(time.Now().Unix() + 60*60*24),
			},
		},
	}
	staff, err := s.addNewStaffToDB(ctx, staff)
	if err != nil {
		return nil, err
	}
	// send email
	err = emails.SendEmail(s.SMTPClient, emails.SendMailConfig{})
	if err != nil {
		return nil, err
	}
	return staff, nil
}

// StaffAcceptInvite verify invite code and set the new staff password and email
// as verified.
func (s Staff) StaffAcceptInvite(ctx context.Context, input *model.StaffAcceptInviteInput) (*models.Staff, error) {
	if input.ConfirmPassword != input.Password {
		return nil, fmt.Errorf("Password and confirm password do not match")
	}
	staff, err := s.fetchStaffFromDB(ctx, input.Email)
	if err != nil {
		return nil, fmt.Errorf("Error fetching staff: %v", err)
	}
	// check if invite code is valid
	if err := validateInviteCode(input.Code, staff.InviteCode); err != nil {
		return nil, err
	}
	// update staff in database
	if err := s.updateStaffInDB(ctx, staff, input); err != nil {
		return nil, fmt.Errorf("Error updating staff: %v", err)
	}
	return staff, nil
}

// StaffUpdate updates the details of the staff i.e. Name, Email, Role.
func StaffUpdate(client *mongo.Client, input *model.StaffUpdateInput) (*models.Staff, error) {
	panic("not implemented")
}

// StaffDelete soft deletes the staff from the database by adding a delatedAt field.
func StaffDelete(client *mongo.Client, input *model.StaffDeleteInput) (*models.Staff, error) {
	panic("not implemented")
}

// StaffChangePassword update the staff password.
func StaffChangePassword(client *mongo.Client, input *model.StaffChangePasswordInput) (*models.Staff, error) {
	panic("not implemented")
}
