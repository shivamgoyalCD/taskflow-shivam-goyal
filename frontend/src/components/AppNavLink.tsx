import { forwardRef } from "react";
import { NavLink, type NavLinkProps } from "react-router-dom";
import { Button, type ButtonProps } from "@mui/material";

type AppNavLinkProps = Omit<ButtonProps<typeof NavLink>, "to"> & {
  to: NavLinkProps["to"];
  label?: string;
};

export const AppNavLink = forwardRef<HTMLAnchorElement, AppNavLinkProps>(
  function AppNavLink({ to, label, children, sx, ...props }, ref) {
    return (
      <Button
        ref={ref}
        component={NavLink}
        to={to}
        color="inherit"
        sx={{
          borderRadius: 999,
          px: 1.5,
          color: "text.secondary",
          "&.active": {
            backgroundColor: "rgba(15, 118, 110, 0.1)",
            color: "primary.main",
          },
          ...sx,
        }}
        {...props}
      >
        {label ?? children}
      </Button>
    );
  },
);
