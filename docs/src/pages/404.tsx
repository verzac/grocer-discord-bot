import * as React from "react";
import { Link } from "gatsby";
import PageContainer from "components/PageContainer";
import { Button, Typography } from "@material-ui/core";
import styled from "styled-components";

const Container = styled.div`
  /* align-items: center; */
  flex: 1 1;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
`;
const NotFoundPage = () => {
  return (
    <PageContainer subtitle="404 Not Found">
      <Container>
        <Typography variant="h1" align="center">
          404 - Nani!?
        </Typography>
        <Typography align="center">
          We can't seem to find what you were looking for :(
        </Typography>
        <Button component={Link} to="/" color="primary">
          Take me back!
        </Button>
      </Container>
    </PageContainer>
  );
};

export default NotFoundPage;
