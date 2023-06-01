import React, { useState, useContext } from "react";
import { AuthContext } from "../contexts/auth-context";
import {
  Button,
  CircularProgress,
  Card,
  CardContent,
  TextField,
  Typography,
} from "@mui/material";

const LoginForm = () => {
  const { login } = useContext(AuthContext);
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [isLoading, setIsLoading] = useState(false);

  const handleUsernameChange = (event) => {
    setUsername(event.target.value);
  };

  const handlePasswordChange = (event) => {
    setPassword(event.target.value);
  };

  const handleSubmit = async (event) => {
    event.preventDefault();
    setIsLoading(true);
    // Simulate a login request with a 1 second delay
    setTimeout(() => {
      setIsLoading(false);
      login("fake_token", false);
    }, 1000);
  };

  return (
    <Card sx={{ maxWidth: 345, margin: "3rem auto" }}>
      <CardContent>
        <form onSubmit={handleSubmit}>
          <Typography variant="h5" gutterBottom>
            Login
          </Typography>
          <TextField
            id="username"
            label="Username"
            variant="outlined"
            value={username}
            onChange={handleUsernameChange}
            required
            margin="normal"
            fullWidth
          />
          <TextField
            id="password"
            label="Password"
            variant="outlined"
            type="password"
            value={password}
            onChange={handlePasswordChange}
            required
            margin="normal"
            fullWidth
          />
          <Button
            variant="contained"
            color="primary"
            type="submit"
            disabled={isLoading}
            startIcon={isLoading ? <CircularProgress size={24} /> : null}
          >
            Login
          </Button>
        </form>
      </CardContent>
    </Card>
  );
};

export default LoginForm;
