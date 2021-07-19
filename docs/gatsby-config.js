module.exports = {
  siteMetadata: {
    siteUrl: "https://www.yourdomain.tld",
    title: "GroceryBot",
  },
  plugins: [
    "gatsby-plugin-react-helmet",
    {
      resolve: `gatsby-theme-material-ui`,
      options: {
        stylesProvider: {
          injectFirst: true,
        },
        webFontsConfig: {
          fonts: {
            google: [
              {
                family: "Montserrat",
                variants: [`300`, `400`, `500`, "600", "700"],
              },
            ],
          },
        },
      },
    },
    "gatsby-plugin-styled-components",
    "gatsby-plugin-tsconfig-paths",
    "gatsby-plugin-image",
    {
      resolve: "gatsby-plugin-sharp",
      options: {
        defaults: {
          placeholder: "none",
        },
      },
    },
    {
      resolve: `gatsby-plugin-manifest`,
      options: {
        name: "GroceryBot | Manage your grocery list.",
        short_name: "GroceryBot",
        start_url: "/",
        background_color: "#3F5E5A",
        theme_color: "#20FC8F",
        // Enables "Add to Homescreen" prompt and disables browser UI (including back button)
        // see https://developers.google.com/web/fundamentals/web-app-manifest/#display
        display: "standalone",
        icon: "src/images/grobot-logo.png", // This path is relative to the root of the site.
        // An optional attribute which provides support for CORS check.
        // If you do not provide a crossOrigin option, it will skip CORS for manifest.
        // Any invalid keyword or empty string defaults to `anonymous`
        // crossOrigin: `use-credentials`,
      },
    },
  ],
};
