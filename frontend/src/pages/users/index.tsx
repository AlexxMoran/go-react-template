import { Alert, Box, CircularProgress, Divider, Paper, Stack, Typography } from "@mui/material";
import { useRootService } from "@shared/hooks/use-root-service";
import { observer } from "mobx-react-lite";
import { useEffect, type FC } from "react";
import { useTranslation } from "react-i18next";

export const UsersPage: FC = observer(() => {
  const { t } = useTranslation();
  const { authService, usersService } = useRootService();
  const { users, isLoading } = usersService;

  useEffect(() => {
    void usersService.loadUsers();
  }, [usersService]);

  return (
    <Stack spacing={3}>
      <Paper sx={{ p: { xs: 3, md: 4 }, borderRadius: 6 }}>
        <Stack spacing={1.5}>
          <Typography variant="h3">{t("protectedPageTitle")}</Typography>
          <Typography color="text.secondary">{t("protectedPageDescription")}</Typography>
          {authService.me && (
            <Alert severity="success">
              {authService.me.email}
              {authService.me.first_name ? ` · ${authService.me.first_name}` : ""}
              {` · ${authService.me.role}`}
            </Alert>
          )}
        </Stack>
      </Paper>

      <Paper sx={{ p: { xs: 3, md: 4 }, borderRadius: 6 }}>
        <Stack spacing={2}>
          <Typography variant="h5">{t("users")}</Typography>
          <Divider />
          {isLoading ? (
            <Box py={4} display="flex" justifyContent="center">
              <CircularProgress />
            </Box>
          ) : users.length === 0 ? (
            <Typography color="text.secondary">{t("noUsers")}</Typography>
          ) : (
            <Stack spacing={2}>
              {users.map((user) => (
                <Paper
                  key={user.id}
                  variant="outlined"
                  sx={{ p: 2.5, borderRadius: 4, backgroundColor: "background.paper" }}
                >
                  <Stack spacing={0.5}>
                    <Typography fontWeight={700}>{user.first_name || user.email}</Typography>
                    <Typography color="text.secondary">{user.email}</Typography>
                    <Typography variant="body2" color="text.secondary">
                      role: {user.role} · verified: {user.is_verified ? "yes" : "no"} · active:{" "}
                      {user.is_active ? "yes" : "no"}
                    </Typography>
                  </Stack>
                </Paper>
              ))}
            </Stack>
          )}
        </Stack>
      </Paper>
    </Stack>
  );
});
