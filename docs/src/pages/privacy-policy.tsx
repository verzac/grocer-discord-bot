import { Link, Typography } from "@material-ui/core";
import PageContainer from "components/PageContainer";
import React from "react";
import FactAccordion from "components/FactAccordion";
import CommandText from "components/CommandText";
import GeneralContentContainer from "components/GeneralContentContainer";

function PrivacyPolicyPage() {
  return (
    <PageContainer subtitle="Privacy Policy">
      <GeneralContentContainer>
        <Typography variant="h1">👀 Privacy policy</Typography>
        <Typography>
          Curious about how your data is handled in GroceryBot? Read on.
        </Typography>
        <Typography variant="h2">TL:DR;</Typography>
        <FactAccordion
          summary={`1. "Maintainers" = people who have full access to the underlying infrastructure that runs GroceryBot. No-one else has access.`}
          detail={
            <Typography>
              Heyo 👋! I&apos;m{" "}
              <Link href="https://github.com/verzac">verzac</Link>, and I&apos;m
              currently the only official maintainer of this bot. The bot runs
              on my AWS server in Singapore, and no-one but the
              &quot;maintainers&quot; have access to the underlying
              infrastructure (including the aforementioned server) that powers
              the bot.
            </Typography>
          }
        />
        <FactAccordion
          summary={`2. Maintainers (who are not in your Discord server) will never look at your grocery list.`}
          detail={
            <>
              <Typography gutterBottom>
                GroceryBot processes your data given by you to the bot through
                its commands so that it can provide you with its intended
                services. Our policy is that NO HUMANS are allowed to manually
                look at GroceryBot&apos;s database, unless if you have
                specifically provided written consent (e.g. to debug an issue
                that you are having with the bot).
              </Typography>
              <Typography gutterBottom>
                It is also important to note that we do aggregate and collect
                anonymised usage data. For example, we record how long
                GroceryBot takes to serve up a request/command within a
                particular period on average. Rest assured though, these data
                are fully anonymous (i.e. from those data alone, no-one can
                figure out who people are and what they&apos;re using the bot
                for).
              </Typography>
            </>
          }
        />
        <FactAccordion
          summary={`3. So how is your data used?`}
          detail={
            <>
              <Typography gutterBottom>
                GroceryBot processes the following data:
              </Typography>
              <Typography gutterBottom>
                <ul>
                  <li>Your grocery entry (duh).</li>
                  <li>
                    Your server ID - this is used to indicate which grocery list
                    is for which server.
                  </li>
                  <li>
                    Your Discord ID - this is a number-based ID from Discord
                    (e.g. 12345678), not your Discord username (e.g.
                    verzac#1234). This is used to power{" "}
                    <CommandText>!grodeets</CommandText> so that you can tell
                    who created what entry.
                  </li>
                  <li>
                    When your grocery entry is updated - removed items are
                    removed immediately and permanently.
                  </li>
                  <li>
                    Your channel ID - this is used so that GroceryBot knows
                    where to send updates to (currently used by{" "}
                    <CommandText>!grohere</CommandText>).
                  </li>
                </ul>
              </Typography>
            </>
          }
        />
        <FactAccordion
          summary={`4. You will always be provided a way to erase your data completely from GroceryBot.`}
          detail={
            <>
              <Typography gutterBottom>
                Pretty much what it says, although exceptions apply to error
                logs, which are always automatically deleted after 14 days.{" "}
              </Typography>
              <Typography gutterBottom>
                While we&apos;re sad to see you go, you can type{" "}
                <CommandText>!groreset</CommandText> to remove ALL of your data
                from GroceryBot (except error logs, which wouldn&apos;t pop up
                if you&apos;ve never encountered an error).
              </Typography>
            </>
          }
        />
      </GeneralContentContainer>
    </PageContainer>
  );
}

export default PrivacyPolicyPage;
