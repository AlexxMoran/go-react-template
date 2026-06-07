import { Layout } from "@app/layout";
import { Pages } from "@app/routes";
import { CssBaseline, CircularProgress, Stack, ThemeProvider } from "@mui/material";
import i18nextInstance from "@shared/config/i18n/config";
import { EAppRoutes } from "@shared/constants/appRoutes";
import { THEME } from "@shared/config/theme";
import { RootServiceContext } from "@shared/hooks/use-root-service";
import { RootService } from "@shared/services/root-service";
import { SnackbarProvider, enqueueSnackbar } from "notistack";
import { useEffect, useState, type FC } from "react";
import { I18nextProvider } from "react-i18next";
import { useNavigate } from "react-router";

export const App: FC = () => {
  const navigate = useNavigate();
  const [isBootstrapping, setIsBootstrapping] = useState(true);
  const [rootService] = useState(
    () =>
      new RootService({
        alertError: (message) => enqueueSnackbar(message, { variant: "error" }),
        redirectToLoginPage: () => navigate(EAppRoutes.Login)
      })
  );

  useEffect(() => {
    void (async () => {
      try {
        await rootService.authService.refreshToken();
      } finally {
        setIsBootstrapping(false);
      }
    })();
  }, [rootService]);

  return (
    <SnackbarProvider maxSnack={3} autoHideDuration={3500}>
      <ThemeProvider theme={THEME}>
        <I18nextProvider i18n={i18nextInstance}>
          <RootServiceContext.Provider value={rootService}>
            <CssBaseline />
            {isBootstrapping ? (
              <Stack minHeight="100vh" alignItems="center" justifyContent="center">
                <CircularProgress />
              </Stack>
            ) : (
              <Layout>
                <Pages />
              </Layout>
            )}
          </RootServiceContext.Provider>
        </I18nextProvider>
      </ThemeProvider>
    </SnackbarProvider>
  );
};
