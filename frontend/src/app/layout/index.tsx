import { AppBar, Box, Button, Container, Stack, Toolbar, Typography } from "@mui/material";
import { EAppRoutes } from "@shared/constants/appRoutes";
import { useRootService } from "@shared/hooks/use-root-service";
import { observer } from "mobx-react-lite";
import type { FC, PropsWithChildren } from "react";
import { useTranslation } from "react-i18next";
import { Link as RouterLink, useNavigate } from "react-router";

export const Layout: FC<PropsWithChildren> = observer(({ children }) => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { authService } = useRootService();

  const handleLogout = async () => {
    await authService.logout();
    navigate(EAppRoutes.Login);
  };

  return (
    <Box minHeight="100vh" sx={{ background: "linear-gradient(180deg, #f4efe7 0%, #fcfaf6 100%)" }}>
      <AppBar position="sticky" color="transparent" elevation={0} sx={{ borderBottom: "1px solid rgba(53, 41, 31, 0.08)", backdropFilter: "blur(8px)" }}>
        <Toolbar sx={{ justifyContent: "space-between", gap: 2 }}>
          <Typography variant="h6" sx={{ fontWeight: 800, letterSpacing: "0.08em" }}>
            GO + REACT TEMPLATE
          </Typography>
          <Stack direction="row" spacing={1}>
            {authService.isAuthenticated ? (
              <>
                <Button component={RouterLink} to={EAppRoutes.Users} color="inherit">
                  {t("users")}
                </Button>
                <Button onClick={handleLogout} variant="contained">
                  {t("logout")}
                </Button>
              </>
            ) : (
              <>
                <Button component={RouterLink} to={EAppRoutes.Login} color="inherit">
                  Login
                </Button>
                <Button component={RouterLink} to={EAppRoutes.Register} variant="contained">
                  Register
                </Button>
              </>
            )}
          </Stack>
        </Toolbar>
      </AppBar>
      <Container maxWidth="lg" sx={{ py: { xs: 3, md: 6 } }}>
        {children}
      </Container>
    </Box>
  );
});
