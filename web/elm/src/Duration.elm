module Duration exposing (..)

import Time exposing (Time)

type alias Duration = Float

between : Time -> Time -> Duration
between a b = b - a

format : Duration -> String
format duration =
  let
    seconds = truncate (duration / 1000)
    remainingSeconds = seconds `rem` 60
    minutes = seconds // 60
    remainingMinutes = minutes `rem` 60
    hours = minutes // 60
    remainingHours = hours `rem` 24
    days = hours // 24
  in
    case (days, remainingHours, remainingMinutes, remainingSeconds) of
      (0, 0, 0, s) ->
        toString s ++ "s"

      (0, 0, m, s) ->
        toString m ++ "m " ++ toString s ++ "s"

      (0, h, m, _) ->
        toString h ++ "h " ++ toString m ++ "m"

      (d, h, _, _) ->
        toString d ++ "d " ++ toString h ++ "h"
