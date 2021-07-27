import { createTheme, responsiveFontSizes } from "@material-ui/core";

const theme = createTheme({
  typography: {
    fontFamily: ["Montserrat", "sans-serif"].join(","),
    h1: {
      fontWeight: 600,
      // color: "black",
    },
    h2: {
      fontWeight: 400,
      letterSpacing: 0.5,
    },
  },
  palette: {
    type: "dark",
    primary: {
      main: "#20FC8F",
    },
    secondary: {
      main: "#3F5E5A",
    },
  },
  overrides: {
    MuiButton: {
      label: {
        textAlign: "center",
      },
    },
    //   MuiTypography: {
    //     caption: {
    //       lineHeight: 0.5,
    //     },
    //   },
  },
});

export default responsiveFontSizes(theme);
