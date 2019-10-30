package auth

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
)

const (
	IssColleague         = "colleague"
	IssMembership        = "membership"
	IssShippingApi       = "shipping-api"
	userClaimContextName = "userClaim"
)

type UserClaim struct {
	Id            int64
	Iss           string
	CustomerId    int64
	ColleagueId   int64
	ShopId        int64
	ChannelId     int64
	TenantCode    string
	Username      string
	ColleagueName string
}

func (UserClaim) FromCtx(ctx context.Context) UserClaim {
	v := ctx.Value(userClaimContextName)
	if v == nil {
		return UserClaim{}
	}

	userClaim, ok := v.(UserClaim)
	if !ok {
		return UserClaim{}
	}

	return userClaim
}

func (UserClaim) FromToken(token string) (UserClaim, error) {
	var userClaim UserClaim

	si := strings.Index(token, ".")
	li := strings.LastIndex(token, ".")
	if si == -1 || li == -1 || si == li {
		return userClaim, errors.New("Invalid token")
	}

	payload := token[si+1 : li]
	if payload == "" {
		return userClaim, errors.New("Invalid token")
	}

	payloadBytes, err := decodeSegment(payload)
	if err != nil {
		return userClaim, err
	}

	if err := json.Unmarshal(payloadBytes, &userClaim); err != nil {
		return userClaim, err
	}

	if userClaim.TenantCode == "" {
		return userClaim, errors.New("Invalid token")
	}
	switch userClaim.Iss {
	case IssColleague:
		userClaim.ColleagueId = userClaim.Id
	case IssMembership:
		userClaim.CustomerId = userClaim.Id
	}

	return userClaim, nil
}
