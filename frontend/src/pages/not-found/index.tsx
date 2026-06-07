import { Box, Button, Paper, Stack, Typography } from "@mui/material";
import { EAppRoutes } from "@shared/constants/appRoutes";
import type { FC } from "react";
import { useTranslation } from "react-i18next";
import { Link } from "react-router";

export const NotFoundPage: FC = () => {
  const { t } = useTranslation();

  return (
    <Box maxWidth={560} mx="auto" pt={{ xs: 4, md: 10 }}>
      <Paper sx={{ p: { xs: 3, md: 5 }, borderRadius: 6 }}>
        <Stack spacing={2}>
          <Typography variant="h3">404</Typography>
          <Typography color="text.secondary">{t("notFound")}</Typography>
          <Button component={Link} to={EAppRoutes.Users} variant="contained">
            {t("users")}
          </Button>
        </Stack>
      </Paper>
    </Box>
  );
};
