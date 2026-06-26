
package models

import (
    "time"
)

type User struct {
    ID                   string     `json:"id"`
    Email                string     `json:"email"`
    Phone                string     `json:"phone"`
    PasswordHash         string     `json:"-"`
    FirstName            string     `json:"first_name"`
    LastName             string     `json:"last_name"`
    DateOfBirth          *time.Time `json:"date_of_birth,omitempty"`
    Role                 string     `json:"role"`
    Status               string     `json:"status"`
    KYCStatus            string     `json:"kyc_status"`
    KYCData              string     `json:"kyc_data,omitempty"`
    MFAEnabled           bool       `json:"mfa_enabled"`
    MFASecret            string     `json:"-"`
    FailedLoginAttempts  int        `json:"-"`
    LockedUntil          *time.Time `json:"locked_until,omitempty"`
    CreatedAt            time.Time  `json:"created_at"`
    UpdatedAt            time.Time  `json:"updated_at"`
    LastLogin            *time.Time `json:"last_login,omitempty"`
    DeletedAt            *time.Time `json:"deleted_at,omitempty"`
}

