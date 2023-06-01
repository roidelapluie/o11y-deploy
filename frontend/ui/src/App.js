import React from "react";
import { BrowserRouter as Router, Route, Routes } from "react-router-dom";
import { Container, Typography } from "@mui/material";
import { ThemeProvider } from "@mui/material/styles";
import CssBaseline from '@mui/material/CssBaseline';

import CustomAppBar from "./components/AppBar";
import Home from "./pages/Home";
import About from "./pages/About";
import Contact from "./pages/Contact";
import Login from "./pages/Login";
import AuthProvider from "./contexts/auth-context"; // Import the auth context

import "./App.css";
import theme from "./theme";

function App() {
  return (
    <Router basename='ui'>
      <ThemeProvider theme={theme}>
      <CssBaseline />
        <AuthProvider>
          <div className="App">
            <CustomAppBar />
            <Container>
              <Routes>
                <Route exact path="/" element={<Home />} />
                <Route path="/about" element={<About />} />
                <Route path="/contact" element={<Contact />} />
                <Route path="/login" element={<Login />} />
              </Routes>
            </Container>
            <footer className="App-footer">
              <Typography variant="body2" color="text.secondary" align="center">
                © 2021 My App. All rights reserved.
              </Typography>
            </footer>
          </div>
        </AuthProvider>
      </ThemeProvider>
    </Router>
  );
}

export default App;
