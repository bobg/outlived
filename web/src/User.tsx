import React from 'react'

import { LogoutButton } from './LogoutButton'
import { ReceiveMailCheckbox } from './ReceiveMailCheckbox'
import { UserData } from './types'

interface Props {
  onLogin: (user: UserData) => void
  user?: UserData
}

interface State {
  born?: string
  email?: string
  password?: string
}

export class User extends React.Component<Props, State> {
  public state: State = {}

  private login = async () => {
    const { email, password } = this.state

    const resp = await fetch('/s/login', {
      method: 'POST',
      credentials: 'same-origin',
      body: JSON.stringify({
        email,
        password,
      }),
    })
    const user = await resp.json() as UserData
    this.props.onLogin(user)
  }

  private signup = () => {
    const { born, email, password } = this.state

    fetch('/s/signup', {
      method: 'POST',
      credentials: 'same-origin',
      body: JSON.stringify({
        email,
        password,
        born,
      }),
    })
  }

  private loginDisabled = (): boolean => {
    return (
      !this.state ||
      !this.state.email ||
      !emailValid(this.state.email) ||
      !this.state.password ||
      !passwordValid(this.state.password)
    )
  }
  private signupDisabled = () => this.loginDisabled()

  public render = () => {
    const { user } = this.props

    if (user) {
      const { csrf, email } = user
      return (
        <div>
          <div>
            Signed in as {email}.
            <LogoutButton csrf={csrf} />
          </div>
          <div>
            <ReceiveMailCheckbox csrf={csrf} user={user} />
          </div>
        </div>
      )
    }

    return (
      <div>
        <label htmlFor='email'>E-mail address</label>
        <input
          type='email'
          id='email'
          name='email'
          onChange={(ev: React.ChangeEvent<HTMLInputElement>) =>
            this.setState({ email: ev.target.value })
          }
        />
        <label htmlFor='password'>Password</label>
        <input
          type='password'
          id='password'
          name='password'
          onChange={(ev: React.ChangeEvent<HTMLInputElement>) =>
            this.setState({ password: ev.target.value })
          }
        />
        <button onClick={() => this.login()} disabled={this.loginDisabled()}>
          Log in
        </button>
        <button onClick={() => this.signup()} disabled={this.signupDisabled()}>
          Sign up
        </button>
      </div>
    )
  }
}

// Adapted from https://www.w3resource.com/javascript/form/email-validation.php.
const emailValid = (inp: string) => {
  return /^\w+([.+-]\w+)*@\w+([.-]?\w+)*(\.\w{2,3})+$/.test(inp)
}

const passwordValid = (inp: string) => {
  return !!inp
}
