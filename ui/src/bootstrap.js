import "assets/style.scss";

import React from "react";
import ReactDOM from "react-dom";

import * as Sentry from "@sentry/browser";

import App from "App";
import { sentryConfig } from "config";
import * as serviceWorker from "serviceWorker";

Sentry.init(sentryConfig);
// Set custom tag 'app', for filtering
Sentry.setTag("app", "xp-ui");

ReactDOM.render(<App />, document.getElementById("root"));

// If you want your app to work offline and load faster, you can change
// unregister() to register() below. Note this comes with some pitfalls.
// Learn more about service workers: https://bit.ly/CRA-PWA
serviceWorker.unregister();
