import React from "react";

interface Props {
  csrf: string;
}

export class LogoutButton extends React.Component<Props> {
  public render = () => (
    <form method="POST" action="/s/logout">
      <input type="hidden" name="csrf" value={this.props.csrf} />
      <button type="submit">Log out</button>
    </form>
  );
}
