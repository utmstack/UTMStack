package templates

const ResetPassword = `<!DOCTYPE html>
<html>
    <head>
        <meta http-equiv="Content-Type" content="text/html; charset=UTF-8" />
        <title>{{.ProductName}} password reset</title>
    </head>
    <body>
        <p>Dear {{.FirstName}},</p>
        <p>A password reset was requested for your {{.ProductName}} account. Please click the link below to set a new password:</p>
        <p><a href="{{.ResetURL}}">{{.ResetURL}}</a></p>
        <p>If you did not request this, you can safely ignore this email.</p>
        <p>Regards,<br/><em>The {{.ProductName}} Team</em></p>
    </body>
</html>
`
