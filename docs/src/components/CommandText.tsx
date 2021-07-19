import { Theme, withTheme } from "@material-ui/core";
import styled from "styled-components";

export default withTheme(styled.span`
  font-weight: 600;
  /* color: ${({ theme }: { theme: Theme }) => theme.palette.primary.main}; */
`);
