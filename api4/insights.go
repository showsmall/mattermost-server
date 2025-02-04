// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package api4

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/mattermost/mattermost-server/v6/model"
)

func (api *API) InitInsights() {
	// Reactions
	api.BaseRoutes.InsightsForTeam.Handle("/reactions", api.APISessionRequired(minimumProfessionalLicense(rejectGuests(getTopReactionsForTeamSince)))).Methods("GET")
	api.BaseRoutes.InsightsForUser.Handle("/reactions", api.APISessionRequired(minimumProfessionalLicense(rejectGuests(getTopReactionsForUserSince)))).Methods("GET")

	// Channels
	api.BaseRoutes.InsightsForTeam.Handle("/channels", api.APISessionRequired(minimumProfessionalLicense(rejectGuests(getTopChannelsForTeamSince)))).Methods("GET")
	api.BaseRoutes.InsightsForUser.Handle("/channels", api.APISessionRequired(minimumProfessionalLicense(rejectGuests(getTopChannelsForUserSince)))).Methods("GET")

	// Threads
	api.BaseRoutes.InsightsForTeam.Handle("/threads", api.APISessionRequired(requireLicense(getTopThreadsForTeamSince))).Methods("GET")
	api.BaseRoutes.InsightsForUser.Handle("/threads", api.APISessionRequired(requireLicense(getTopThreadsForUserSince))).Methods("GET")

	// New teammembers
	api.BaseRoutes.InsightsForTeam.Handle("/team_members", api.APISessionRequired(minimumProfessionalLicense(rejectGuests(getNewTeamMembersSince)))).Methods("GET")
}

// Top Reactions

func getTopReactionsForTeamSince(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireTeamId()
	if c.Err != nil {
		return
	}

	team, appErr := c.App.GetTeam(c.Params.TeamId)
	if appErr != nil {
		c.Err = appErr
		return
	}

	if !c.App.SessionHasPermissionToTeam(*c.AppContext.Session(), team.Id, model.PermissionViewTeam) {
		c.SetPermissionError(model.PermissionViewTeam)
		return
	}

	user, appErr := c.App.GetUser(c.AppContext.Session().UserId)
	if appErr != nil {
		c.Err = appErr
		return
	}

	startTime := model.StartOfDayForTimeRange(c.Params.TimeRange, user.GetTimezoneLocation())

	topReactionList, appErr := c.App.GetTopReactionsForTeamSince(c.Params.TeamId, c.AppContext.Session().UserId, &model.InsightsOpts{
		StartUnixMilli: startTime.UnixMilli(),
		Page:           c.Params.Page,
		PerPage:        c.Params.PerPage,
	})
	if appErr != nil {
		c.Err = appErr
		return
	}

	js, err := json.Marshal(topReactionList)
	if err != nil {
		c.Err = model.NewAppError("getTopReactionsForTeamSince", "api.marshal_error", nil, "", http.StatusInternalServerError).Wrap(err)
		return
	}

	w.Write(js)
}

func getTopReactionsForUserSince(c *Context, w http.ResponseWriter, r *http.Request) {
	c.Params.TeamId = r.URL.Query().Get("team_id")

	// TeamId is an optional parameter
	if c.Params.TeamId != "" {
		if !model.IsValidId(c.Params.TeamId) {
			c.SetInvalidURLParam("team_id")
			return
		}

		team, appErr := c.App.GetTeam(c.Params.TeamId)
		if appErr != nil {
			c.Err = appErr
			return
		}

		if !c.App.SessionHasPermissionToTeam(*c.AppContext.Session(), team.Id, model.PermissionViewTeam) {
			c.SetPermissionError(model.PermissionViewTeam)
			return
		}
	}

	user, appErr := c.App.GetUser(c.AppContext.Session().UserId)
	if appErr != nil {
		c.Err = appErr
		return
	}

	startTime := model.StartOfDayForTimeRange(c.Params.TimeRange, user.GetTimezoneLocation())

	topReactionList, appErr := c.App.GetTopReactionsForUserSince(c.AppContext.Session().UserId, c.Params.TeamId, &model.InsightsOpts{
		StartUnixMilli: startTime.UnixMilli(),
		Page:           c.Params.Page,
		PerPage:        c.Params.PerPage,
	})
	if appErr != nil {
		c.Err = appErr
		return
	}

	js, err := json.Marshal(topReactionList)
	if err != nil {
		c.Err = model.NewAppError("getTopReactionsForUserSince", "api.marshal_error", nil, "", http.StatusInternalServerError).Wrap(err)
		return
	}

	w.Write(js)
}

// Top Channels

func getTopChannelsForTeamSince(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireTeamId()
	if c.Err != nil {
		return
	}

	team, appErr := c.App.GetTeam(c.Params.TeamId)
	if appErr != nil {
		c.Err = appErr
		return
	}

	if !c.App.SessionHasPermissionToTeam(*c.AppContext.Session(), team.Id, model.PermissionViewTeam) {
		c.SetPermissionError(model.PermissionViewTeam)
		return
	}

	user, appErr := c.App.GetUser(c.AppContext.Session().UserId)
	if appErr != nil {
		c.Err = appErr
		return
	}

	loc := user.GetTimezoneLocation()
	startTime := model.StartOfDayForTimeRange(c.Params.TimeRange, loc)

	topChannels, appErr := c.App.GetTopChannelsForTeamSince(c.AppContext, c.Params.TeamId, c.AppContext.Session().UserId, &model.InsightsOpts{
		StartUnixMilli: startTime.UnixMilli(),
		Page:           c.Params.Page,
		PerPage:        c.Params.PerPage,
	})
	if appErr != nil {
		c.Err = appErr
		return
	}

	topChannels.PostCountByDuration, appErr = postCountByDurationViewModel(c, topChannels, startTime, c.Params.TimeRange, nil, loc)
	if appErr != nil {
		c.Err = appErr
		return
	}

	js, err := json.Marshal(topChannels)
	if err != nil {
		c.Err = model.NewAppError("getTopChannelsForTeamSince", "api.marshal_error", nil, "", http.StatusInternalServerError).Wrap(err)
		return
	}

	w.Write(js)
}

func getTopChannelsForUserSince(c *Context, w http.ResponseWriter, r *http.Request) {
	c.Params.TeamId = r.URL.Query().Get("team_id")

	// TeamId is an optional parameter
	if c.Params.TeamId != "" {
		if !model.IsValidId(c.Params.TeamId) {
			c.SetInvalidURLParam("team_id")
			return
		}

		team, appErr := c.App.GetTeam(c.Params.TeamId)
		if appErr != nil {
			c.Err = appErr
			return
		}

		if !c.App.SessionHasPermissionToTeam(*c.AppContext.Session(), team.Id, model.PermissionViewTeam) {
			c.SetPermissionError(model.PermissionViewTeam)
			return
		}
	}

	user, appErr := c.App.GetUser(c.AppContext.Session().UserId)
	if appErr != nil {
		c.Err = appErr
		return
	}

	loc := user.GetTimezoneLocation()
	startTime := model.StartOfDayForTimeRange(c.Params.TimeRange, loc)

	topChannels, appErr := c.App.GetTopChannelsForUserSince(c.AppContext, c.AppContext.Session().UserId, c.Params.TeamId, &model.InsightsOpts{
		StartUnixMilli: startTime.UnixMilli(),
		Page:           c.Params.Page,
		PerPage:        c.Params.PerPage,
	})
	if appErr != nil {
		c.Err = appErr
		return
	}

	topChannels.PostCountByDuration, appErr = postCountByDurationViewModel(c, topChannels, startTime, c.Params.TimeRange, &c.AppContext.Session().UserId, loc)
	if appErr != nil {
		c.Err = appErr
		return
	}

	js, err := json.Marshal(topChannels)
	if err != nil {
		c.Err = model.NewAppError("getTopChannelsForUserSince", "api.marshal_error", nil, "", http.StatusInternalServerError).Wrap(err)
		return
	}

	w.Write(js)
}

// Top Threads
func getTopThreadsForTeamSince(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireTeamId()
	if c.Err != nil {
		return
	}

	team, appErr := c.App.GetTeam(c.Params.TeamId)
	if appErr != nil {
		c.Err = appErr
		return
	}

	// license check
	lic := c.App.Srv().License()
	if lic.SkuShortName != model.LicenseShortSkuProfessional && lic.SkuShortName != model.LicenseShortSkuEnterprise {
		c.Err = model.NewAppError("", "api.insights.license_error", nil, "", http.StatusNotImplemented)
		return
	}

	// restrict guests and users with no access to team
	user, appErr := c.App.GetUser(c.AppContext.Session().UserId)
	if appErr != nil {
		c.Err = appErr
		return
	}

	if !c.App.SessionHasPermissionToTeam(*c.AppContext.Session(), team.Id, model.PermissionViewTeam) || user.IsGuest() {
		c.SetPermissionError(model.PermissionViewTeam)
		return
	}

	startTime := model.StartOfDayForTimeRange(c.Params.TimeRange, user.GetTimezoneLocation())

	topThreads, appErr := c.App.GetTopThreadsForTeamSince(c.AppContext, c.Params.TeamId, c.AppContext.Session().UserId, &model.InsightsOpts{
		StartUnixMilli: startTime.UnixMilli(),
		Page:           c.Params.Page,
		PerPage:        c.Params.PerPage,
	})
	if appErr != nil {
		c.Err = appErr
		return
	}

	js, err := json.Marshal(topThreads)
	if err != nil {
		c.Err = model.NewAppError("getTopThreadsForTeamSince", "api.marshal_error", nil, "", http.StatusInternalServerError).Wrap(err)
		return
	}

	w.Write(js)
}

func getTopThreadsForUserSince(c *Context, w http.ResponseWriter, r *http.Request) {
	c.Params.TeamId = r.URL.Query().Get("team_id")

	// restrict guests and users with no access to team
	user, appErr := c.App.GetUser(c.AppContext.Session().UserId)
	if appErr != nil {
		c.Err = appErr
		return
	}
	// TeamId is an optional parameter
	if c.Params.TeamId != "" {
		if !model.IsValidId(c.Params.TeamId) {
			c.SetInvalidURLParam("team_id")
			return
		}

		team, teamErr := c.App.GetTeam(c.Params.TeamId)
		if teamErr != nil {
			c.Err = teamErr
			return
		}

		// license check
		lic := c.App.Srv().License()
		if lic.SkuShortName != model.LicenseShortSkuProfessional && lic.SkuShortName != model.LicenseShortSkuEnterprise {
			c.Err = model.NewAppError("", "api.insights.license_error", nil, "", http.StatusNotImplemented)
			return
		}

		if !c.App.SessionHasPermissionToTeam(*c.AppContext.Session(), team.Id, model.PermissionViewTeam) || user.IsGuest() {
			c.SetPermissionError(model.PermissionViewTeam)
			return
		}
	}

	startTime := model.StartOfDayForTimeRange(c.Params.TimeRange, user.GetTimezoneLocation())

	topThreads, appErr := c.App.GetTopThreadsForUserSince(c.AppContext, c.Params.TeamId, c.AppContext.Session().UserId, &model.InsightsOpts{
		StartUnixMilli: startTime.UnixMilli(),
		Page:           c.Params.Page,
		PerPage:        c.Params.PerPage,
	})
	if appErr != nil {
		c.Err = appErr
		return
	}

	js, err := json.Marshal(topThreads)
	if err != nil {
		c.Err = model.NewAppError("getTopThreadsForUserSince", "api.marshal_error", nil, "", http.StatusInternalServerError).Wrap(err)
		return
	}

	w.Write(js)
}

// postCountByDurationViewModel expects a list of channels that are pre-authorized for the given user to view.
func postCountByDurationViewModel(c *Context, topChannelList *model.TopChannelList, startTime *time.Time, timeRange string, userID *string, location *time.Location) (model.ChannelPostCountByDuration, *model.AppError) {
	if len(topChannelList.Items) == 0 {
		return nil, nil
	}
	var postCountsByDay []*model.DurationPostCount
	channelIDs := topChannelList.ChannelIDs()
	var grouping model.PostCountGrouping
	if timeRange == model.TimeRangeToday {
		grouping = model.PostsByHour
	} else {
		grouping = model.PostsByDay
	}
	postCountsByDay, err := c.App.PostCountsByDuration(c.AppContext, channelIDs, startTime.UnixMilli(), userID, grouping, location)
	if err != nil {
		return nil, err
	}
	return model.ToDailyPostCountViewModel(postCountsByDay, startTime, model.TimeRangeToNumberDays(timeRange), channelIDs), nil
}

func getNewTeamMembersSince(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireTeamId()
	if c.Err != nil {
		return
	}

	team, err := c.App.GetTeam(c.Params.TeamId)
	if err != nil {
		c.Err = err
		return
	}

	if !c.App.SessionHasPermissionToTeam(*c.AppContext.Session(), team.Id, model.PermissionViewTeam) {
		c.SetPermissionError(model.PermissionViewTeam)
		return
	}

	user, err := c.App.GetUser(c.AppContext.Session().UserId)
	if err != nil {
		c.Err = err
		return
	}
	loc := user.GetTimezoneLocation()
	startTime := model.StartOfDayForTimeRange(c.Params.TimeRange, loc)

	ntms, count, err := c.App.GetNewTeamMembersSince(c.AppContext, c.Params.TeamId, &model.InsightsOpts{
		StartUnixMilli: startTime.UnixMilli(),
		Page:           c.Params.Page,
		PerPage:        c.Params.PerPage,
	})
	if err != nil {
		c.Err = err
		return
	}

	ntms.TotalCount = count

	js, jsonErr := json.Marshal(ntms)
	if jsonErr != nil {
		c.Err = model.NewAppError("getNewTeamembersForTeamSince", "api.marshal_error", nil, jsonErr.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(js)
}
