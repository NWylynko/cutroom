# Friction report

- **Outcome:** done. Reset the inherited edit, rebuilt it, rendered, and `check` passed with 0 warnings.
- **Commands run:** 9 — guide; guide workflow/edit/verify + cat of json files; status/transcript/transcript --suspects/timeline; edit reset; edit tighten; edit cut (retake); edit cut (aside); timeline + timeline --joins; render; check (several were chained in single shell calls).
- **Errors hit:** none.
- **Guessed wrong:** nothing. I did not read the transcript of the final render (`transcript --file`), only the check output.
- **Confusing output:** a chained command's output printed in odd order (timeline warnings before the tighten summary). Otherwise clear.
- **Missing:** nothing essential. Judgement call: I cut the "By the way, my cat..." aside as a digression; the tool only flags it, it does not say whether the task wants it kept.
- **Not sure it's right:** the inherited edit had a cut inside a sentence ("I'm | to"), a kept retake with "UM", a kept aside and long pauses. Reset plus tighten plus two cuts fixed all of them. I could not listen to the result, only trust `check` (words, stray-sound, sync all ok).
- **Best part:** `timeline` warnings printed ready-made cut commands, and `edit reset` made fixing a wrongly built edit one step.
- **One change:** have the `status` next-step suggest `edit reset` when several warnings are open, since "undo" for broken-sentence alone would not have fixed the rest.
