package api

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/smtp"

	"github.com/Leagueify/api/internal/model"
	"github.com/Leagueify/api/internal/util"
	"github.com/labstack/echo/v4"
)

func (api *API) Email(e *echo.Group) {
	e.POST("/email/config", api.requiresAdmin(api.createConfig))
}

func (api *API) createConfig(c echo.Context) error {
	var emailConfig model.EmailConfig

	if err := c.Bind(&emailConfig); err != nil {
		return util.SendStatus(http.StatusBadRequest, c, "invalid json payload")
	}

	if err := c.Validate(emailConfig); err != nil {
		return util.SendStatus(http.StatusBadRequest, c, util.HandleError(err))
	}

	// attempt to validate credentials
	tlsConfig := &tls.Config{InsecureSkipVerify: false, ServerName: emailConfig.SMTPHost}
	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%v", emailConfig.SMTPHost, emailConfig.SMTPPort), tlsConfig)
	if err != nil {
		return util.SendStatus(http.StatusNotFound, c, "host not found")
	}
	client, err := smtp.NewClient(conn, emailConfig.SMTPHost)
	if err != nil {
		return util.SendStatus(http.StatusBadRequest, c, util.HandleError(err))
	}
	auth := smtp.PlainAuth("", emailConfig.SMTPUser, emailConfig.SMTPPass, emailConfig.SMTPHost)
	if err := client.Auth(auth); err != nil {
		return util.SendStatus(http.StatusUnauthorized, c, util.HandleError(err))
	}
	if err := client.Quit(); err != nil {
		return util.SendStatus(http.StatusBadRequest, c, util.HandleError(err))
	}

	// check for existing configs
	totalEmailConfigs, err := api.DB.GetTotalEmailConfigs()
	if err != nil {
		return util.SendStatus(http.StatusBadRequest, c, util.HandleError(err))
	}
	if totalEmailConfigs != 0 {
		return util.SendStatus(http.StatusUnauthorized, c, "")
	}

	emailConfig.ID = util.SignedToken(4)
	emailConfig.IsEnabled = true
	if err := api.DB.CreateEmailConfig(emailConfig); err != nil {
		return util.SendStatus(http.StatusBadRequest, c, util.HandleError(err))
	}
	return c.JSON(http.StatusCreated,
		map[string]string{
			"status": "successful",
		},
	)
}
