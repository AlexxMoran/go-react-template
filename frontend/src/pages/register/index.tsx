import { Box, Button, Paper, Stack, TextField, Typography } from "@mui/material";
import { EAppRoutes } from "@shared/constants/appRoutes";
import { useRootService } from "@shared/hooks/use-root-service";
import { Form, Formik } from "formik";
import { useSnackbar } from "notistack";
import type { FC } from "react";
import { useTranslation } from "react-i18next";
import { Link, useNavigate } from "react-router";
import * as Yup from "yup";

export const RegisterPage: FC = () => {
  const { t } = useTranslation();
  const { enqueueSnackbar } = useSnackbar();
  const navigate = useNavigate();
  const { authApiService } = useRootService();

  const validationSchema = Yup.object({
    email: Yup.string().email().required(),
    password: Yup.string().min(8).required(),
    first_name: Yup.string().max(50).optional()
  });

  return (
    <Box maxWidth={520} mx="auto" pt={{ xs: 4, md: 8 }}>
      <Paper sx={{ p: { xs: 3, md: 5 }, borderRadius: 6 }}>
        <Stack spacing={3}>
          <Stack spacing={1}>
            <Typography variant="h3">{t("register")}</Typography>
            <Typography color="text.secondary">{t("subtitle")}</Typography>
          </Stack>

          <Formik
            initialValues={{ email: "", password: "", first_name: "" }}
            validationSchema={validationSchema}
            onSubmit={async (values) => {
              await authApiService.register(values);
              enqueueSnackbar(t("registrationSuccess"), { variant: "success" });
              navigate(EAppRoutes.Login);
            }}
          >
            {({ errors, touched, values, handleChange, isSubmitting }) => (
              <Form>
                <Stack spacing={2.5}>
                  <TextField
                    name="email"
                    label={t("email")}
                    value={values.email}
                    onChange={handleChange}
                    error={touched.email && !!errors.email}
                    helperText={touched.email ? errors.email : ""}
                    fullWidth
                  />
                  <TextField
                    name="first_name"
                    label={t("firstName")}
                    value={values.first_name}
                    onChange={handleChange}
                    error={touched.first_name && !!errors.first_name}
                    helperText={touched.first_name ? errors.first_name : ""}
                    fullWidth
                  />
                  <TextField
                    name="password"
                    label={t("password")}
                    type="password"
                    value={values.password}
                    onChange={handleChange}
                    error={touched.password && !!errors.password}
                    helperText={touched.password ? errors.password : ""}
                    fullWidth
                  />
                  <Button type="submit" variant="contained" size="large" disabled={isSubmitting}>
                    {t("register")}
                  </Button>
                </Stack>
              </Form>
            )}
          </Formik>

          <Typography color="text.secondary">
            <Link to={EAppRoutes.Login}>{t("backToLogin")}</Link>
          </Typography>
        </Stack>
      </Paper>
    </Box>
  );
};
