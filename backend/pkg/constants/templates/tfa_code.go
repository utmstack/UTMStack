package templates

const TfaCode = `
<!DOCTYPE html>
<html>
    <head>
        <title>{{.ProductName}} two-factor verification code</title>
        <meta http-equiv="Content-Type" content="text/html; charset=UTF-8" />
    </head>
    <body>
        <p>Dear {{.FirstName}},</p>
        <p>Your {{.ProductName}} two-factor verification code is:</p>
        <p style="font-size:22px;font-weight:bold;letter-spacing:4px;">{{.Code}}</p>
        <p>This code expires shortly. If you did not request it, you can safely ignore this email.</p>
        <p>Regards,<br/><em>{{.ProductName}}</em></p>
    </body>
</html>
`
