import {
  AccordionSummary,
  Typography,
  AccordionDetails,
} from "@material-ui/core";
import React from "react";
import ControlledAccordion from "./ControlledAccordion";
import ExpandMoreIcon from "@material-ui/icons/ExpandMore";

interface FactAccordionProps {
  summary: React.ReactNode;
  detail: React.ReactNode;
}

function FactAccordion({ summary, detail }: FactAccordionProps) {
  return (
    <ControlledAccordion>
      <AccordionSummary expandIcon={<ExpandMoreIcon />}>
        <Typography variant="h6">{summary}</Typography>
      </AccordionSummary>
      <AccordionDetails>
        <div>{detail}</div>
      </AccordionDetails>
    </ControlledAccordion>
  );
}

export default FactAccordion;
