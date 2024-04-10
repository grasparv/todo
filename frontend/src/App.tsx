import React, { useState } from 'react';
import './App.css';
import '@fontsource/roboto/300.css';
import Button from '@mui/material/Button';
import TextField from '@mui/material/TextField';
import Home from './pages/Home';

const App: React.FC = () => {
    return (
        <Login />
    );
}

const Login: React.FC = () => {
    const [user, setUser] = useState<string>("");
    const [done, setDone] = useState<boolean>(false);

    const validateUser = () => {
        if (user === undefined || user.length < 1) {
            alert("Please fill in a username");
            return false;
        }

        return true;
    }

    const handleChangeEvent = (event: React.ChangeEvent<HTMLInputElement>) => {
        setUser(event.target.value);
    }

    const finishOnPage = () => {
        if (!validateUser()) {
            return;
        }
        setDone(true);
    }

    const handleKeyDown = (event: React.KeyboardEvent<HTMLInputElement>) => {
        if (event.key === "Enter") {
            finishOnPage();
        }
    }

    const handleOnClick = (event: React.MouseEvent<HTMLButtonElement>) => {
        finishOnPage();
    }

    if (done !== true) {
        return (
            <div className="login">
              <div className="nameInput">
                <TextField
                  autoFocus
                  label="Your name"
                  value={user}
                  onChange={(event) => handleChangeEvent(event)}
                  onKeyDown={(event) => handleKeyDown(event)}
                  />
              </div>
              <div className="nameInput">
                <Button variant="contained" onClick={handleOnClick}>Login</Button>
              </div>
            </div>
        );
    }

    return (
         <Home user={user} />
    );
}

export default App;
