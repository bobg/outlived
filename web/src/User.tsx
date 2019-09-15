import React from "react";

import { LogoutButton } from "./LogoutButton";
import { ReceiveMailCheckbox } from "./ReceiveMailCheckbox";

interface Props {
  csrf: string;
  user: any; // xxx
}

interface State {
  email?: string;
  password?: string;
}

export class User extends React.Component<Props, State> {
  private login = () => {
    const { email, password } = this.state;

    fetch("xxx", {
      method: "POST",
      credentials: "same-origin",
      body: JSON.stringify({
        email,
        password
      })
    });
  };

  private signup = () => {
    // xxx
  };

  private loginDisabled = (): boolean => {
    return (
      !this.state.email ||
      !emailValid(this.state.email) ||
      !this.state.password ||
      !passwordValid(this.state.password)
    );
  };
  private signupDisabled = () => this.loginDisabled();

  public render = () => {
    const { csrf, user } = this.props;

    if (user) {
      return (
        <div>
          <div>
            Signed in as {user.email}.
            <LogoutButton csrf={csrf} />
          </div>
          <div>
            <ReceiveMailCheckbox csrf={csrf} user={user} />
          </div>
        </div>
      );
    }

    return (
      <div>
        <label htmlFor="email">E-mail address</label>
        <input
          type="email"
          id="email"
          name="email"
          onChange={(ev: React.ChangeEvent<HTMLInputElement>) =>
            this.setState({ email: ev.target.value })
          }
        />
        <label htmlFor="password">Password</label>
        <input
          type="password"
          id="password"
          name="password"
          onChange={(ev: React.ChangeEvent<HTMLInputElement>) =>
            this.setState({ password: ev.target.value })
          }
        />
        <button onClick={() => this.login} disabled={this.loginDisabled()}>
          Log in
        </button>
        <button onClick={() => this.signup} disabled={this.signupDisabled()}>
          Sign up
        </button>
      </div>
    );
  };
}

// Adapted from https://www.w3resource.com/javascript/form/email-validation.php.
const emailValid = (inp: string) => {
  return /^\w+([.+-]\w+)*@\w+([.-]?\w+)*(\.\w{2,3})+$/.test(inp);
};

const passwordValid = (inp: string) => {
  return !!inp;
};
