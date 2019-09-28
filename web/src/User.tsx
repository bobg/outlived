import React from 'react'
import { Modal, ModalBody, ModalTitle } from 'react-bootstrap'
import DatePicker from 'react-date-picker'

import { LogoutButton } from './LogoutButton'
import { PasswordDialog } from './Password'
import { post } from './post'
import { UserData } from './types'
import { tzname } from './tz'

interface Props {
  onLogin: (user: UserData) => void
  user?: UserData
}

interface State {
  birthDate: Date
  email?: string
  enteringBirthdate?: boolean
  enteringPassword?: boolean
  password?: string
  receivingMail: boolean
  signingUp?: boolean
}

export class User extends React.Component<Props, State> {
  constructor(props: Props) {
    super(props)
    this.state = {
      birthDate: new Date('1969-7-20'),
      receivingMail: !!props.user && props.user.verified && props.user.active,
    }
  }

  private login = async () => {
    const { email, password } = this.state

    const resp = await post('/s/login', {
      email,
      password,
      tzname: tzname(),
    })
    const user = (await resp.json()) as UserData
    this.props.onLogin(user)
  }

  private signup = () => {
    const { birthDate, email, password } = this.state

    post('/s/signup', {
      email,
      password,
      born: birthDate, // xxx
      tzname: tzname(),
    })
  }

  private setActive = (active: boolean) => {
    if (!this.props.user) {
      return
    }
    const { csrf } = this.props.user
    const resp = post('/s/setactive', {
      csrf,
      active,
    })
    // xxx check resp
    this.setState({ receivingMail: active })
  }

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

  public render = () => {
    const { user } = this.props

    if (user) {
      const { csrf, email } = user
      return (
        <div>
          <div>
            Logged in as {email}.
            <LogoutButton csrf={csrf} />
          </div>
          <div>
            <label htmlFor='active'>Receive Outlived mail?</label>
            <input
              type='checkbox'
              id='active'
              name='active'
              checked={this.state.receivingMail}
              disabled={!user.verified}
              onChange={ev => this.setActive(ev.target.checked)}
            />
          </div>
        </div>
      )
    }

    return (
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
            <DatePicker
              onChange={(date: Date | Date[]) => {
                this.setState({ birthDate: date as Date }) // xxx hack
              }}
              value={this.state.birthDate}
            />
          </ModalBody>
        </Modal>
      </>
    )
  }
}

// Adapted from https://www.w3resource.com/javascript/form/email-validation.php.
const emailValid = (inp?: string) => {
  return inp && /^\w+([.+-]\w+)*@\w+([.-]?\w+)*(\.\w{2,3})+$/.test(inp)
}
