import React from 'react'
import { Modal, ModalBody, ModalTitle } from 'react-bootstrap'
import DatePicker from 'react-date-picker'
import Toggle from 'react-toggle'

import { LogoutButton } from './LogoutButton'
import { PasswordDialog } from './Password'
import { post } from './post'
import { UserData } from './types'
import { tzname } from './tz'

interface LoggedInProps {
  user: UserData
}

interface LoggedInState {
  receivingMail: boolean
  reverified: boolean
}

export class LoggedInUser extends React.Component<
  LoggedInProps,
  LoggedInState
> {
  constructor(props: LoggedInProps) {
    super(props)
    const { user } = props
    this.state = {
      receivingMail: user.verified && user.active,
      reverified: false,
    }
  }

  private setActive = async (active: boolean) => {
    const { csrf } = this.props.user
    const resp = await post('/s/setactive', {
      csrf,
      active,
    })
    // xxx check resp
    this.setState({ receivingMail: active })
  }

  private reverify = async () => {
    const { csrf } = this.props.user
    const resp = await post('/s/reverify', { csrf })
    // xxx check resp
    this.setState({ reverified: true })
  }

  public render = () => {
    const { user } = this.props
    const { csrf, email, verified } = user

    return (
      <div className='user'>
        <div>
          Logged in as {email}.
          <LogoutButton csrf={csrf} />
        </div>
        <div>
          <label htmlFor='active'>
            <span>Receive Outlived mail?</span>
            <Toggle
              id='active'
              checked={!!this.state.receivingMail}
              disabled={!user.verified}
              onChange={ev => this.setActive(ev.target.checked)}
            />
          </label>
          {!verified &&
            (this.state.reverified ? (
              <span id='reverified'>
                Check your e-mail for a verification message from Outlived.
              </span>
            ) : (
              <span id='unconfirmed'>
                You have not yet confirmed your e-mail address.
                <button id='resend-button' onClick={this.reverify}>
                  Resend verification
                </button>
              </span>
            ))}
        </div>
      </div>
    )
  }
}

interface LoggedOutProps {
  onLogin: (user: UserData) => void
}

interface LoggedOutState {
  birthDate: Date
  email?: string
  enteringBirthdate?: boolean
  enteringPassword?: boolean
  password?: string
  signingUp?: boolean
}

export class LoggedOutUser extends React.Component<
  LoggedOutProps,
  LoggedOutState
> {
  public state: LoggedOutState = { birthDate: new Date() }

  private onPassword = async (pw: string) => {
    const { email } = this.state
    this.setState({
      enteringBirthdate: this.state.signingUp,
      password: pw,
    })
    if (this.state.signingUp) {
      return
    }
    const resp = await post('/s/login', {
      email,
      password: pw,
      tzname: tzname(),
    })
    // xxx check for error
    const user = (await resp.json()) as UserData
    this.props.onLogin(user)
  }

  private onBirthDate = async (ev: React.FormEvent<HTMLFormElement>) => {
    ev.preventDefault()
    const { birthDate } = this.state
    if (!birthdateValid(birthDate)) {
      return
    }
    const { email, password } = this.state
    const resp = await post('/s/signup', {
      email,
      password,
      born: datestr(birthDate),
      tzname: tzname(),
    })
    const user = (await resp.json()) as UserData
    this.props.onLogin(user)
  }

  public render = () => (
    <>
      <p>Log in to see whom you’ve recently outlived.</p>

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
        <button
          onClick={() => this.setState({ enteringPassword: true })}
          disabled={!emailValid(this.state.email)}
        >
          Log in
        </button>
        <button
          onClick={() =>
            this.setState({ enteringPassword: true, signingUp: true })
          }
          disabled={!emailValid(this.state.email)}
        >
          Sign up
        </button>
      </div>

      <PasswordDialog
        prompt={this.state.signingUp ? 'Choose password' : 'Enter password'}
        show={() => !!this.state.enteringPassword}
        onClose={() => this.setState({ enteringPassword: false })}
        onSubmit={this.onPassword}
      />

      <Modal
        onHide={() =>
          this.setState({ enteringBirthdate: false, signingUp: false })
        }
        show={this.state.enteringBirthdate}
      >
        <Modal.Header closeButton>
          <ModalTitle>Birth date</ModalTitle>
        </Modal.Header>
        <ModalBody>
          <form onSubmit={this.onBirthDate}>
            <DatePicker
              onChange={(d: Date | Date[]) => {
                const date = d as Date // xxx hack
                this.setState({ birthDate: date })
              }}
              value={this.state.birthDate}
            />
            <button
              type='submit'
              disabled={!birthdateValid(this.state.birthDate)}
            >
              Submit
            </button>
          </form>
        </ModalBody>
      </Modal>
    </>
  )
}

// Adapted from https://www.w3resource.com/javascript/form/email-validation.php.
const emailValid = (inp?: string) => {
  return inp && /^\w+([.+-]\w+)*@\w+([.-]?\w+)*(\.\w{2,3})+$/.test(inp)
}

const yearMillis =
  365 /* days */ * 24 /* hours */ * 60 /* mins */ * 60 /* secs */ * 1000

const birthdateValid = (d: Date) => {
  const now = new Date()
  const diff = now.valueOf() - d.valueOf()
  if (diff < 13 * yearMillis) {
    return false
  }
  if (diff > 200 * yearMillis) {
    return false
  }
  return true
}

const datestr = (d: Date) => {
  const y = d.getFullYear()
  const m = 1 + d.getMonth() // come on javascript
  const day = d.getDate() // come ON, javascript
  return y + '-' + ('0' + m).substr(-2) + '-' + ('0' + day).substr(-2)
}
