import React from 'react';

interface Props {
  csrf: string
  user: any // xxx
}

class User extends React.Component<xxx, Props> {
  public render() {
    {user ? (
      <div>
        Signed in as {user.email}.
        <LogoutButton csrf={csrf} />
      </div>
      <div>
        <ReceiveMailCheckbox csrf={csrf} user={user} />
      </div>
    ) : (
      <label for="email">
        E-mail address
      </label>
      <input type="email" id="email" name="email" />
      <label for="password">
        Password
      </label>
      <input type="password" id="password" name="password" />
    )}
  }
}
