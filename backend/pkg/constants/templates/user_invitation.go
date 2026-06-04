package templates

const UserInvitation = `<!DOCTYPE html>
<html>
    <head>
        <meta http-equiv="Content-Type" content="text/html; charset=UTF-8" />
        <title>UTMStack account created</title>
    </head>
    <body>
        <p>Dear {{.FirstName}},</p>
        <p>Your UTMStack account has been created. Please click the link below to set your password and activate your account:</p>
        <p><a href="{{.InviteURL}}">{{.InviteURL}}</a></p>
        <p>Regards,<br/><em>The UTMStack Team</em></p>
    </body>
</html>
`
