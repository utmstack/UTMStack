package templates

const NginxCustomBadGateway string = `<!DOCTYPE html>
<html lang="es">
<head>
  <meta charset="UTF-8">
  <title>UTMStack - Maintenance</title>
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <style>
    @import url('https://fonts.googleapis.com/css2?family=Roboto:wght@400;700&display=swap');

    * {
      margin: 0;
      padding: 0;
      box-sizing: border-box;
    }

    body {
      font-family: 'Roboto', sans-serif;
      background: url("https://storage.googleapis.com/utmstack-updates/nginx/login-background.jpg") no-repeat center center fixed;
      background-size: cover;
      color: #333;
      display: flex;
      justify-content: center;
      align-items: center;
      min-height: 100vh;
      text-align: center;
      padding: 20px;
    }

    .container {
      max-width: 600px;
      background-color: white;
      padding: 40px;
      border-radius: 15px;
      backdrop-filter: blur(10px);
      box-shadow: 0 8px 32px rgba(0, 0, 0, 0.3);
      animation: fadeIn 1.5s ease-out;
    }
    
    .icon img {
      width: 80px;
      max-width: 100%;
      height: auto;
      object-fit: contain;
    }

    h1 {
      font-size: 48px;
      margin-bottom: 20px;
    }

    p {
      font-size: 20px;
      line-height: 1.5;
      margin-bottom: 30px;
    }

    .icon {
      font-size: 60px;
      margin-bottom: 20px;
    }

    .footer {
      font-size: 14px;
      margin-top: 30px;
      color: #999;
    }

    @keyframes fadeIn {
      from {
        opacity: 0;
        transform: translateY(-20px);
      }
      to {
        opacity: 1;
        transform: translateY(0);
      }
    }
  </style>
</head>
<body>
  <div class="container">
    <div class="icon">
      <img src="https://storage.googleapis.com/utmstack-updates/nginx/logo_UTMStack.svg"
    </div>
    <h1>We're getting things ready</h1>
    <p>
      <strong>UTMStack</strong> is currently under maintenance.<br>Please check back in a few minutes.
    </p>
    <div class="footer">
      &copy; UTMStack ©Copyright 2025. All Rights Reserved.
    </div>
  </div>
</body>
</html>
`
