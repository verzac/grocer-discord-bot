// @ts-check
// Note: type annotations allow type checking and IDEs autocompletion

const lightCodeTheme = require("prism-react-renderer/themes/github");
const darkCodeTheme = require("prism-react-renderer/themes/dracula");

/** @type {import('@docusaurus/types').Config} */
const config = {
  title: "GroceryBot",
  tagline: "Get a grocery list going for your Discord server!",
  url: "https://grocerybot.net",
  baseUrl: "/",
  onBrokenLinks: "throw",
  onBrokenMarkdownLinks: "warn",
  favicon: "img/grocerybot.png",
  organizationName: "verzac", // Usually your GitHub org/user name.
  projectName: "grocer-discord-bot", // Usually your repo name.

  presets: [
    [
      "@docusaurus/preset-classic",
      /** @type {import('@docusaurus/preset-classic').Options} */
      ({
        docs: {
          sidebarPath: require.resolve("./sidebars.js"),
          // Please change this to your repo.
          editUrl: "https://github.com/facebook/docusaurus/edit/main/website/",
        },
        blog: {
          showReadingTime: true,
          // Please change this to your repo.
          // editUrl:
          //   "https://github.com/facebook/docusaurus/edit/main/website/blog/",
        },
        theme: {
          customCss: require.resolve("./src/css/custom.css"),
        },
        gtag: {
          trackingID: "G-LEYDSTJJ7L",
          anonymizeIP: true,
        },
      }),
    ],
    [
      "redocusaurus",
      {
        // Plugin Options for loading OpenAPI files
        specs: [
          {
            spec: "../openapi.yaml",
            route: "/api/",
          },
        ],
      },
    ],
  ],

  themeConfig:
    /** @type {import('@docusaurus/preset-classic').ThemeConfig} */
    ({
      colorMode: {
        defaultMode: "dark",
        disableSwitch: true,
      },
      navbar: {
        title: "GroceryBot",
        logo: {
          alt: "GroceryBot Logo",
          src: "img/grocerybot.png",
        },
        items: [
          {
            type: "doc",
            docId: "intro",
            position: "left",
            label: "Tutorial",
          },
          { to: "/blog", label: "Blog", position: "left" },
          { to: "/api", label: "API", position: "left" },
          {
            href: "https://github.com/verzac/grocer-discord-bot",
            label: "GitHub",
            position: "right",
          },
        ],
      },
      footer: {
        style: "dark",
        links: [
          {
            title: "Docs",
            items: [
              {
                label: "Tutorial",
                to: "/docs/intro",
              },
              {
                label: "Privacy Policy",
                to: "/privacy-policy",
              },
            ],
          },
          {
            title: "Community",
            items: [
              {
                label: "Invite GroceryBot to Your Server",
                href: "https://discord.com/oauth2/authorize?client_id=815120759680532510&permissions=2048&scope=bot%20applications.commands",
              },
              {
                label: "Discord Support Server",
                href: "https://discord.com/invite/rBjUaZyskg",
              },
            ],
          },
          {
            title: "More",
            items: [
              // {
              //   label: "Blog",
              //   to: "/blog",
              // },
              {
                label: "GitHub",
                href: "https://github.com/verzac/grocer-discord-bot",
              },
            ],
          },
        ],
        copyright: `Copyright © ${new Date().getFullYear()} GroceryBot. Built with Docusaurus.`,
      },
      prism: {
        theme: lightCodeTheme,
        darkTheme: darkCodeTheme,
      },
    }),
};

module.exports = config;
