const { ModuleFederationPlugin } = require("webpack").container;
const paths = require("react-scripts/config/paths");
const deps = require("./package.json").dependencies;
const appName = "xp";

module.exports = ({ }) => ({
    plugins: [
        {
            plugin: {
                // Background:
                // https://github.com/facebook/create-react-app/issues/9510#issuecomment-902536147
                overrideWebpackConfig: ({ webpackConfig }) => {
                    // Set auto for Module Federation to work
                    webpackConfig.output.publicPath = "auto";
                    // Set custom publicPath
                    const htmlWebpackPlugin = webpackConfig.plugins.find(
                        (plugin) => plugin.constructor.name === "HtmlWebpackPlugin"
                    );
                    htmlWebpackPlugin.userOptions = {
                        ...htmlWebpackPlugin.userOptions,
                        publicPath: paths.publicUrlOrPath,
                        // Exclude the exposed app for hot reloading to work
                        excludeChunks: [appName],
                    };

                    return webpackConfig;
                },
            },
        },
    ],
    webpack: {
        plugins: {
            add: [
                new ModuleFederationPlugin({
                    name: appName,
                    exposes: {
                        "./EditExperimentEngineConfig": "./src/turing/components/form/EditExperimentEngineConfig",
                        "./ExperimentEngineConfigDetails": "./src/turing/components/configuration/ExperimentEngineConfigDetails",
                        "./ExperimentsLandingPage": "./src/experiments/ExperimentsLandingPage",
                    },
                    filename: "remoteEntry.js",
                    shared: {
                        ...deps,
                        react: {
                            shareScope: "default",
                            singleton: true,
                            requiredVersion: deps.react,
                        },
                        "react-dom": {
                            singleton: true,
                            requiredVersion: deps["react-dom"],
                        },
                        /* 
                        Without singleton declaration, 2 different versions of @gojek/mlp-ui dependency were loaded
                        which caused the parent app to crash due to different "global states" being used.
                        
                        In addition, due to the nature of dynamic import, Module Federation is not able to use the 
                        singleton dependency's version specified in XP, even if compatible with the dependency version
                        on host. Hence, the MLP dependency version that matters in the remote loading scenario is only
                        what is specified in the parent app (Turing).
                        */
                        "@gojek/mlp-ui": {
                            singleton: true,
                            requiredVersion: deps["@gojek/mlp-ui"],
                        }
                    },
                }),
            ],
        }
    },
});
