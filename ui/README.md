# XP UI

This is the UI component of XP based on React web framework.

## Prerequisites

- [Node.js](https://nodejs.org/en/download/) v14
- yarn 1.22.x (`npm install --global yarn`)

## Install Dependencies

To start developing, you need to first download the packages required by XP UI.
```bash
yarn install
```

## Configuration

The React application can be configured using environment variables. By default, 
it will read from config values from `.env` file. Refer to 
this [documentation](https://create-react-app.dev/docs/adding-custom-environment-variables/) for more details.

### Config Variables
| Environment Variable            | Required | Default  | Description     |
| ------------------------------- | -------- | -------- | --------------- |
| `REACT_APP_API_TIMEOUT`         | No       | 5000     | Timeout (in milliseconds) for requests to API |
| `REACT_APP_XP_API`              | Yes      |          | Endpoint of XP API | 
| `REACT_APP_MLP_API`             | Yes      |          | Endpoint of MLP API | 
| ...                             | ...      | ...      | ... | 
| `REACT_APP_SENTRY_DSN`          | No       |          | Sets the upper bound of the number max replicas that user can configure for router/enricher/ensembler | 

### Setup Google OAuth2

XP UI currently depends on Google OAuth2 to determine the identity of the
users. In order to use XP UI, you need to setup Google OAuth2 and set 
the client id for your web application.

1. Open [Google API Console Credentials page](https://console.developers.google.com/apis/credentials)
2. Use an existing Google project or create a new project
3. Select **+ Create Credentials**, then **OAuth client ID**
4. You may be prompted to set a product name on the Consent screen; if so, click **Configure consent screen**, supply the requested information, and click **Save** to return to the Credentials screen
5. Select **Web Application** as the **Application Type** and enter any additional information required.
6. For the **Authorized JavaScript origins**, make sure the enter the following URLs for local devlopment: 
   `http://localhost:3001`. This is the default URL when starting XP UI React app in development mode.
7. Click **Create**
8. Copy the **client ID** value shown on the next page. You will need this value to configure XP UI authentication.
9. Set the value for `REACT_APP_OAUTH_CLIENT_ID` in `.env.development` file with the client ID you just received.

For more info and, please refer to this Google [documentation](https://developers.google.com/identity/protocols/oauth2/javascript-implicit-flow).

## Start Local Development Server

Run the following command to start the app in development mode.
```bash
yarn start
```

Open [http://localhost:3001](http://localhost:3001) to view the app in the browser.

## Build for Production

Run the following command to build XP UI app for production. 
It correctly bundles React in production mode and optimizes the build for the best performance.
```bash
yarn build
```

The final build will be available under the `/build` folder, which is ready to
be deployed. Refer to the [deployment](https://facebook.github.io/create-react-app/docs/deployment) for more information.
