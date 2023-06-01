import React, { useState, useContext } from "react";
import { AppBar, Box, Toolbar, Typography, Button, Link } from "@mui/material";
import { Menu as MenuIcon } from "@mui/icons-material";
import { AuthContext } from "../contexts/auth-context";

function CustomAppBar() {
  const { authState, logout } = useContext(AuthContext);
  return (
    <Box sx={{ flexGrow: 1 }}>
      <AppBar position="static">
        <Toolbar>
          <Typography variant="h6" component="div" sx={{ flexGrow: 1, fontWeight: 900, fontFamily: 'Montserrat' }}>
            O11y
          </Typography>
      {!authState.token &&
          <Button color="inherit" component={Link} to="/login">Login</Button>
          ||
          <Button color="inherit" onClick={logout}>Logout</Button>
      }
        </Toolbar>
      </AppBar>
    </Box>
  );
}

export default CustomAppBar;
