package helpers

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strconv"
	"strings"

	"github.com/user/video-downloader-backend/internal/dto"
	"github.com/user/video-downloader-backend/internal/model"
	"github.com/user/video-downloader-backend/internal/repository"
)

type MailHelper interface {
	SendResetPasswordEmail(ctx context.Context, email, resetToken string) error
	SendContactEmail(ctx context.Context, payload *dto.ContactRequest) error
}

type mailHelper struct {
	settingService repository.SettingRepository
}

func NewMailHelper(settingRepo repository.SettingRepository) MailHelper {
	return &mailHelper{settingService: settingRepo}
}

func (m *mailHelper) SendResetPasswordEmail(ctx context.Context, email, resetToken string) error {
	setting, err := m.GetPublicSettings(ctx)
	if err != nil {
		return err
	}

	settingEmail := setting.EMAIL
	if !settingEmail.SMTPEnabled {
		return fmt.Errorf("SMTP is disabled in settings")
	}

	// Validate required SMTP settings
	if settingEmail.SMTPHost == "" || settingEmail.SMTPPort == 0 || settingEmail.SMTPUser == "" || settingEmail.SMTPPassword == "" {
		return fmt.Errorf("incomplete SMTP configuration")
	}

	resetURL := fmt.Sprintf("%s/reset-password?token=%s", strings.TrimRight(setting.WEBSITE.SiteURL, "/"), resetToken)
	siteName := setting.WEBSITE.SiteName
	if siteName == "" {
		siteName = "Video Downloader"
	}

	subject := fmt.Sprintf("Reset Password - %s", siteName)
	body := m.getResetPasswordHTML(siteName, resetURL)

	headers := make(map[string]string)
	headers["From"] = fmt.Sprintf("%s <%s>", settingEmail.FromName, settingEmail.FromEmail)
	headers["To"] = email
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=\"UTF-8\""

	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body

	// Setup SMTP Auth
	auth := smtp.PlainAuth("", settingEmail.SMTPUser, settingEmail.SMTPPassword, settingEmail.SMTPHost)

	// SMTP Server Address
	addr := fmt.Sprintf("%s:%d", settingEmail.SMTPHost, settingEmail.SMTPPort)

	// If port is 465, use implicit TLS
	if settingEmail.SMTPPort == 465 {
		// Handle TLS config for non-localhost
		tlsConfig := &tls.Config{
			InsecureSkipVerify: false,
			ServerName:         settingEmail.SMTPHost,
		}

		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return fmt.Errorf("failed to dial TLS: %w", err)
		}

		c, err := smtp.NewClient(conn, settingEmail.SMTPHost)
		if err != nil {
			return fmt.Errorf("failed to create SMTP client: %w", err)
		}
		defer c.Quit()

		if err = c.Auth(auth); err != nil {
			return fmt.Errorf("failed to auth: %w", err)
		}

		if err = c.Mail(settingEmail.FromEmail); err != nil {
			return fmt.Errorf("failed to set sender: %w", err)
		}

		if err = c.Rcpt(email); err != nil {
			return fmt.Errorf("failed to set recipient: %w", err)
		}

		w, err := c.Data()
		if err != nil {
			return fmt.Errorf("failed to create data writer: %w", err)
		}

		_, err = w.Write([]byte(message))
		if err != nil {
			return fmt.Errorf("failed to write body: %w", err)
		}

		err = w.Close()
		if err != nil {
			return fmt.Errorf("failed to close data writer: %w", err)
		}

		return nil

	} else {
		// Standard smtp.SendMail for port 587 (STARTTLS) or 25 (Plain)
		err = smtp.SendMail(addr, auth, settingEmail.FromEmail, []string{email}, []byte(message))
		if err != nil {
			return fmt.Errorf("failed to send email: %w", err)
		}
	}

	return nil
}

func (m *mailHelper) SendContactEmail(ctx context.Context, payload *dto.ContactRequest) error {
	setting, err := m.GetPublicSettings(ctx)
	if err != nil {
		return err
	}

	settingEmail := setting.EMAIL
	if !settingEmail.SMTPEnabled {
		return fmt.Errorf("SMTP is disabled in settings")
	}

	// Validate required SMTP settings
	if settingEmail.SMTPHost == "" || settingEmail.SMTPPort == 0 || settingEmail.SMTPUser == "" || settingEmail.SMTPPassword == "" {
		return fmt.Errorf("incomplete SMTP configuration")
	}

	siteName := setting.WEBSITE.SiteName
	if siteName == "" {
		siteName = "Video Downloader"
	}

	siteUrl := setting.WEBSITE.SiteURL
	if siteUrl == "" {
		siteUrl = "https://video-downloader.com"
	}

	subject := fmt.Sprintf("Contact Us - %s", siteName)
	body := m.getContactHTML(siteName, siteUrl, payload)

	headers := make(map[string]string)
	headers["From"] = fmt.Sprintf("%s <%s>", settingEmail.FromName, settingEmail.FromEmail)
	headers["To"] = setting.WEBSITE.SiteEmail
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=\"UTF-8\""

	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body

	// Setup SMTP Auth
	auth := smtp.PlainAuth("", settingEmail.SMTPUser, settingEmail.SMTPPassword, settingEmail.SMTPHost)

	// SMTP Server Address
	addr := fmt.Sprintf("%s:%d", settingEmail.SMTPHost, settingEmail.SMTPPort)

	// If port is 465, use implicit TLS
	if settingEmail.SMTPPort == 465 {
		// Handle TLS config for non-localhost
		tlsConfig := &tls.Config{
			InsecureSkipVerify: false,
			ServerName:         settingEmail.SMTPHost,
		}

		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return fmt.Errorf("failed to dial TLS: %w", err)
		}

		c, err := smtp.NewClient(conn, settingEmail.SMTPHost)
		if err != nil {
			return fmt.Errorf("failed to create SMTP client: %w", err)
		}
		defer c.Quit()

		if err = c.Auth(auth); err != nil {
			return fmt.Errorf("failed to auth: %w", err)
		}

		if err = c.Mail(settingEmail.FromEmail); err != nil {
			return fmt.Errorf("failed to set sender: %w", err)
		}

		if err = c.Rcpt(setting.WEBSITE.SiteEmail); err != nil {
			return fmt.Errorf("failed to set recipient: %w", err)
		}

		w, err := c.Data()
		if err != nil {
			return fmt.Errorf("failed to create data writer: %w", err)
		}

		_, err = w.Write([]byte(message))
		if err != nil {
			return fmt.Errorf("failed to write body: %w", err)
		}

		err = w.Close()
		if err != nil {
			return fmt.Errorf("failed to close data writer: %w", err)
		}

		return nil

	} else {
		// Standard smtp.SendMail for port 587 (STARTTLS) or 25 (Plain)
		err = smtp.SendMail(addr, auth, settingEmail.FromEmail, []string{setting.WEBSITE.SiteEmail}, []byte(message))
		if err != nil {
			return fmt.Errorf("failed to send email: %w", err)
		}
	}

	return nil
}

func (m *mailHelper) getResetPasswordHTML(siteName, resetURL string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; background-color: #f9f9f9; }
        .header { background-color: #007bff; color: white; padding: 20px; text-align: center; border-radius: 5px 5px 0 0; }
        .content { background-color: white; padding: 30px; border-radius: 0 0 5px 5px; box-shadow: 0 2px 5px rgba(0,0,0,0.1); }
        .button { display: inline-block; padding: 12px 24px; background-color: #007bff; color: white; text-decoration: none; border-radius: 4px; margin-top: 20px; font-weight: bold; }
        .footer { text-align: center; margin-top: 20px; font-size: 12px; color: #666; }
        p { margin-bottom: 15px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1 style="margin:0;">%s</h1>
        </div>
        <div class="content">
            <h2>Reset Your Password</h2>
            <p>Hello,</p>
            <p>We received a request to reset your password. If you didn't make this request, you can safely ignore this email.</p>
            <p>To reset your password, click the button below:</p>
            <div style="text-align: center;">
                <a href="%s" class="button">Reset Password</a>
            </div>
            <p style="margin-top: 30px; font-size: 14px;">Or copy and paste this link into your browser:</p>
            <p style="font-size: 13px; color: #007bff; word-break: break-all;"><a href="%s">%s</a></p>
            <p>This link will expire in 1 hour.</p>
        </div>
        <div class="footer">
            <p>&copy; %d %s. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`, siteName, resetURL, resetURL, resetURL, 2025, siteName) // Hardcoded year for simplicity or use time.Now().Year()
}

func (m *mailHelper) getContactHTML(siteName, siteURL string, payload *dto.ContactRequest) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; background-color: #f9f9f9; }
        .header { background-color: #007bff; color: white; padding: 20px; text-align: center; border-radius: 5px 5px 0 0; }
        .content { background-color: white; padding: 30px; border-radius: 0 0 5px 5px; box-shadow: 0 2px 5px rgba(0,0,0,0.1); }
        .button { display: inline-block; padding: 12px 24px; background-color: #007bff; color: white; text-decoration: none; border-radius: 4px; margin-top: 20px; font-weight: bold; }
        .footer { text-align: center; margin-top: 20px; font-size: 12px; color: #666; }
        p { margin-bottom: 15px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1 style="margin:0;">%s</h1>
        </div>
        <div class="content">
            <h2>New Contact Request</h2>
            <p>You received a new contact request from %s.</p>
            <p>Email: %s</p>
            <p>Name: %s</p>
            <p>Message: %s</p>
            <p style="margin-top: 30px; font-size: 14px;">Or copy and paste this link into your browser:</p>
            <p style="font-size: 13px; color: #007bff; word-break: break-all;"><a href="%s">%s</a></p>
            <p>This link will expire in 1 hour.</p>
        </div>
        <div class="footer">
            <p>&copy; %d %s. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`, siteName, payload.Name, payload.Email, payload.Name, payload.Message, siteURL, siteURL, 2025, siteName)
}

func (s *mailHelper) GetPublicSettings(ctx context.Context) (*model.SettingsResponse, error) {
	settings, err := s.settingService.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	response := &model.SettingsResponse{}

	for _, item := range settings {
		switch item.GroupName {
		case "WEBSITE":
			mapWebsiteSetting(&response.WEBSITE, item)
		case "EMAIL":
			mapEmailSetting(&response.EMAIL, item)
		case "SYSTEM":
			mapSystemSetting(&response.SYSTEM, item)
		}
	}

	return response, nil
}

func mapWebsiteSetting(target *model.SettingWeb, s model.Setting) {
	switch s.Key {
	case "site_name":
		target.SiteName = s.Value
	case "site_tagline":
		target.SiteTagline = s.Value
	case "site_description":
		target.SiteDescription = s.Value
	case "site_keywords":
		target.SiteKeywords = s.Value
	case "site_logo":
		target.SiteLogo = s.Value
	case "site_favicon":
		target.SiteFavicon = s.Value
	case "site_email":
		target.SiteEmail = s.Value
	case "site_phone":
		target.SitePhone = s.Value
	case "site_url":
		target.SiteURL = s.Value
	}
}

func mapEmailSetting(target *model.SettingEmail, s model.Setting) {
	switch s.Key {
	case "smtp_enabled":
		target.SMTPEnabled = s.Value == "true"
	case "smtp_service":
		target.SMTPService = s.Value
	case "smtp_host":
		target.SMTPHost = s.Value
	case "smtp_port":
		if port, err := strconv.Atoi(s.Value); err == nil {
			target.SMTPPort = port
		}
	case "smtp_user":
		target.SMTPUser = s.Value
	case "smtp_password":
		target.SMTPPassword = s.Value
	case "from_email":
		target.FromEmail = s.Value
	case "from_name":
		target.FromName = s.Value
	}
}

func mapSystemSetting(target *model.SettingSystem, s model.Setting) {
	switch s.Key {
	case "maintenance_mode":
		target.MaintenanceMode = s.Value == "true"
	case "maintenance_message":
		target.MaintenanceMessage = s.Value
	case "source_logo_favicon":
		target.SourceLogoFavicon = s.Value
	}
}
