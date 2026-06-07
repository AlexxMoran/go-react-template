import { LoginPage } from "@pages/login";
import { NotFoundPage } from "@pages/not-found";
import { RegisterPage } from "@pages/register";
import { UsersPage } from "@pages/users";
import { EAppRoutes } from "@shared/constants/appRoutes";
import withAuth from "@shared/hocs/with-auth";
import type { FC } from "react";
import { Navigate, Route, Routes } from "react-router";

const UsersPageWithAuth = withAuth(UsersPage);

export const Pages: FC = () => {
  return (
    <Routes>
      <Route path="/" element={<Navigate to={EAppRoutes.Users} replace />} />
      <Route path={EAppRoutes.Login} element={<LoginPage />} />
      <Route path={EAppRoutes.Register} element={<RegisterPage />} />
      <Route path={EAppRoutes.Users} element={<UsersPageWithAuth />} />
      <Route path="*" element={<NotFoundPage />} />
    </Routes>
  );
};
