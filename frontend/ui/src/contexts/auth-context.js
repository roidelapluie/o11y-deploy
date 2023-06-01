import React, { createContext, useState, useEffect } from "react";
import { useCookies } from "react-cookie";
import { useNavigate } from "react-router-dom";

export const AuthContext = createContext();

const AuthProvider = (props) => {
  const [cookies, setCookie] = useCookies(["sessionId"]);
  const [isRedirecting, setIsRedirecting] = useState(false);
  const [authState, setAuthState] = useState({
    token: cookies.sessionId ?? null,
    isPermanent: false,
  });

  const login = (token, isPermanent) => {
    setAuthState({
      token,
      isPermanent,
    });
    // Set the session ID cookie with the provided options
    setCookie("sessionId", token, {
      // If isPermanent is true, expire the cookie in 1 year
      maxAge: isPermanent ? 365 * 24 * 60 * 60 : undefined,
      // Set the sameSite and secure flags for better security
      sameSite: "strict",
      secure: true,
    });
        navigate('/')
  };

  const logout = () => {
    setAuthState({ token: null, isPermanent: false });
    setCookie("sessionId", null, {
      // Remove the cookie by setting maxAge to 0
      maxAge: 0,
      sameSite: "strict",
      secure: true,
    });
  };

    useEffect(() => {
          setIsRedirecting(false);
    }, [authState.token]);

    const navigate = useNavigate();
    if (!isRedirecting && !authState.token) {
        setIsRedirecting(true);
        navigate('/login')
    }

  const authContextValue = {
    authState,
    login,
    logout,
  };

  return (
    <AuthContext.Provider value={authContextValue}>
      {props.children}
    </AuthContext.Provider>
  );
};

export default AuthProvider;
