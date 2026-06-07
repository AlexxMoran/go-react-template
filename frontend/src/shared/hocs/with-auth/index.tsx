import { Alert, Box } from "@mui/material";
import { EAppRoutes } from "@shared/constants/appRoutes";
import { useRootService } from "@shared/hooks/use-root-service";
import type { ComponentType, FC, PropsWithChildren } from "react";
import { Navigate } from "react-router";

function withAuth<P extends object>(WrappedComponent: ComponentType<P>): FC<PropsWithChildren<P>> {
  const ProtectedComponent: FC<P> = (props) => {
    const { authService } = useRootService();

    if (!authService.isAuthenticated) {
      return (
        <Box maxWidth={560} mx="auto">
          <Alert severity="warning" sx={{ mb: 2 }}>
            Authentication required
          </Alert>
          <Navigate to={EAppRoutes.Login} replace />
        </Box>
      );
    }

    return <WrappedComponent {...props} />;
  };

  return ProtectedComponent;
}

export default withAuth;
