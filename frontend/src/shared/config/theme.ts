import { createTheme } from "@mui/material";

export const THEME = createTheme({
  palette: {
    primary: {
      main: "#9c4f24",
      contrastText: "#fffaf5"
    },
    secondary: {
      main: "#2b6f63"
    },
    background: {
      default: "#fcfaf6",
      paper: "#fffdf9"
    },
    text: {
      primary: "#2f241c",
      secondary: "#6e5b4b"
    }
  },
  shape: {
    borderRadius: 18
  },
  typography: {
    fontFamily: '"Avenir Next", "Montserrat", "Segoe UI", sans-serif',
    h3: {
      fontWeight: 800,
      lineHeight: 1.05
    },
    h4: {
      fontWeight: 800
    },
    h5: {
      fontWeight: 700
    },
    button: {
      textTransform: "none",
      fontWeight: 700
    }
  },
  components: {
    MuiPaper: {
      styleOverrides: {
        root: {
          border: "1px solid rgba(53, 41, 31, 0.08)",
          boxShadow: "0 24px 64px rgba(88, 63, 42, 0.08)"
        }
      }
    },
    MuiButton: {
      styleOverrides: {
        root: {
          borderRadius: 999
        }
      }
    }
  }
});
