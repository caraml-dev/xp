import "assets/style.scss";

import React from "react";
import { BrowserRouter } from "react-router-dom";

import * as Sentry from "@sentry/browser";

import App from "App";
import * as serviceWorker from "serviceWorker";

import { ConfigProvider, useConfig } from "./config";
import { createRoot } from "react-dom/client";

const SentryApp = ({ children }) => {
  const {
    sentryConfig: { dsn, environment, tags },
  } = useConfig();

  Sentry.init({ dsn, environment });
  Sentry.setTags(tags);

  return children;
};

const XPUI = () => (
  <React.StrictMode>
    <ConfigProvider>
      <SentryApp>
        <BrowserRouter>
          <App />
        </BrowserRouter>
      </SentryApp>
    </ConfigProvider>
  </React.StrictMode>
);

const container = document.getElementById("root")
const root = createRoot(container);

root.render(
    XPUI()
);


// If you want your app to work offline and load faster, you can change
// unregister() to register() below. Note this comes with some pitfalls.
// Learn more about service workers: https://bit.ly/CRA-PWA
serviceWorker.unregister();
