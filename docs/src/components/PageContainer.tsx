import {
  Button,
  CssBaseline,
  Paper,
  Typography,
  Link as MuiLink,
  Divider,
} from "@material-ui/core";
import React, { ReactChild } from "react";
import styled from "styled-components";
import { Helmet } from "react-helmet";
import "./layout.css";
import { Link } from "gatsby";
import defaultImage from "images/grobot-logo.png";

const RootContainer = styled.main`
  padding: 24px;
  padding-top: 64px;
  /* height: 200%; */
  min-width: 100%;
  flex: 1 0 100vh;
  /* overflow: visible; */
  display: flex;
  flex-direction: column;
  justify-content: flex-end;
  & > * {
    margin-bottom: 24px;
  }
  & > :last-child {
    margin-bottom: 0;
  }
`;

const AppBar = styled(Paper)`
  display: flex;
  position: fixed;
  flex-direction: row;
  justify-content: flex-end;
  border-radius: 0;
  padding: 8px 16px 8px 16px;
  width: 100%;
  backdrop-filter: blur(20px);
  box-shadow: 0 0 20px rgba(0, 0, 0, 0.5);
  background: rgba(66, 66, 66, 0.5);
  z-index: 100;
`;

const FooterDivider = styled(Divider)`
  margin-bottom: 8px;
`;

interface PageContainerProps {
  children: ReactChild;
  title?: string;
  description?: string;
  className?: string;
  subtitle?: string;
  image?: string;
}

function PageContainer({
  children,
  className,
  title = "GroceryBot",
  subtitle,
  description = "Use Discord to manage your grocery list. No more switching apps just to talk about which grocery item you should get.",
  image = defaultImage,
}: PageContainerProps) {
  return (
    <>
      <Helmet>
        <title>{[title, subtitle].filter(Boolean).join(" | ")}</title>
        <meta name="description" content={description} />
        <meta name="image" content={image} />
      </Helmet>
      <AppBar>
        <Button component={Link} to="/" variant="text" color="primary">
          Home
        </Button>
        <Button variant="text" color="primary" disabled>
          Docs (Coming soon!)
        </Button>
      </AppBar>
      <RootContainer className={className}>
        <CssBaseline />
        {children}
        <div>
          <FooterDivider />
          <Typography align="center">
            GroceryBot and this website is maintained by{" "}
            <MuiLink
              href="https://github.com/verzac"
              target="_blank"
              rel="noopener"
            >
              this guy
            </MuiLink>
            .
          </Typography>
        </div>
      </RootContainer>
    </>
  );
}

export default PageContainer;
