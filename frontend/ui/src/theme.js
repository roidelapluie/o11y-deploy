import { createTheme } from "@mui/material/styles";

const theme = createTheme({
  palette: {
    primary: {
      main: "#FF0048",
      light: "#FF5B72",
      dark: "#C9002C",
    },
    secondary: {
      main: "#FFB900",
      light: "#FFE259",
      dark: "#C78B00",
    },
    background: {
      default: "#dddddd",
    },
    surface: {
      main: "#F7F7F7",
    },
    error: {
      main: "#D32F2F",
    },
    warning: {
      main: "#FFC107",
    },
    info: {
      main: "#1976D2",
    },
    success: {
      main: "#4CAF50",
    },
  },
});

export default theme;
