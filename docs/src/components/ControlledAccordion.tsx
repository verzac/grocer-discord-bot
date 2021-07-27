import { Accordion, AccordionProps } from "@material-ui/core";
import React, { useState } from "react";

function ControlledAccordion(props: AccordionProps) {
  const [open, setOpen] = useState(false);
  function handleChange() {
    setOpen((o) => !o);
  }
  return <Accordion {...props} expanded={open} onChange={handleChange} />;
}

export default ControlledAccordion;
