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
import { StaticImage } from "gatsby-plugin-image";
import useSiteMetadata from "hooks/useSiteMetadata";
import { useLocation } from "@reach/router";

interface RootContainerProps {
  noDefaultFlex?: boolean;
}

const RootContainer = styled.main<RootContainerProps>`
  padding: 24px;
  padding-top: 64px;
  /* height: 200%; */
  min-width: 100%;
  flex: 1 0 100vh;
  /* overflow: visible; */
  display: flex;
  flex-direction: column;
  justify-content: flex-end;
  align-items: center;
  & > * {
    margin-bottom: 24px;
  }
  & > :last-child {
    margin-bottom: 0;
  }
  ${({ noDefaultFlex }) =>
    !noDefaultFlex &&
    `& > :first-child {
    flex: 1 1;
  }`}
`;

const AppBar = styled(Paper)`
  display: flex;
  position: fixed;
  flex-direction: row;
  justify-content: flex-end;
  align-items: center;
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

const FooterArea = styled.div`
  display: flex;
  flex-direction: row;
  justify-content: center;
  align-items: center;
  & > * {
    margin-right: 8px;
  }
  & > :last-child {
    margin-right: 0;
  }
`;

const LogoContainer = styled(Link)`
  flex: 1 1;
`;

interface PageContainerProps {
  children: ReactChild;
  title?: string;
  description?: string;
  className?: string;
  subtitle?: string;
  image?: string;
  rootContainerProps?: RootContainerProps;
}

function PageContainer({
  children,
  className,
  title = "GroceryBot",
  subtitle,
  description = "Use Discord to manage your grocery list. No more switching apps just to talk about which grocery item you should get.",
  image = defaultImage,
  rootContainerProps = {},
}: PageContainerProps) {
  const location = useLocation();
  const { siteUrl } = useSiteMetadata();
  const pageTitle = [title, subtitle].filter(Boolean).join(" | ");
  return (
    <>
      <Helmet>
        <title>{pageTitle}</title>
        <meta name="description" content={description} />
        <meta name="image" content={image} />
        <meta name="og:description" content={description} />
        <meta name="og:title" content={pageTitle} />
        <meta name="og:url" content={`${siteUrl}${location.pathname}`} />
        <meta name="og:image" content={`${siteUrl}${image}`} />
      </Helmet>
      <AppBar>
        <LogoContainer to="/">
          <StaticImage
            src="../images/grobot-logo.png"
            alt="GroceryBot Logo"
            aspectRatio={1}
            width={28}
          />
        </LogoContainer>
        <Button component={Link} to="/" variant="text" color="primary">
          Home
        </Button>
        <Button component={Link} to="/docs" variant="text" color="primary">
          Docs
        </Button>
      </AppBar>
      <RootContainer className={className} {...rootContainerProps}>
        <CssBaseline />
        {children}
        <div>
          <FooterDivider />
          <Typography align="center" gutterBottom>
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
          <FooterArea>
            <Button
              component={Link}
              to="/privacy-policy"
              variant="text"
              color="primary"
            >
              Our Privacy Policy
            </Button>
            <Divider orientation="vertical" flexItem />
            <Button
              component={MuiLink}
              target="_blank"
              rel="noopener"
              href="https://discord.com/oauth2/authorize?client_id=815120759680532510&permissions=2048&scope=bot"
              color="primary"
            >
              Invite to Discord
            </Button>
          </FooterArea>
        </div>
      </RootContainer>
    </>
  );
}

export default PageContainer;
