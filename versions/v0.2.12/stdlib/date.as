static spawn Months = [
  "January",
  "February",
  "March",
  "April",
  "May",
  "June",
  "July",
  "August",
  "September",
  "October",
  "November",
  "December"
];


$ *
$ * new (): Date;
$ * new (value: number | string): Date;
$ * new (year: number, monthIndex: number, date?: number, hours?: number, minutes?: number, seconds?: number, ms?: number): Date;
$ *
$ * Creates a new Date.
$ * @param year The full year designation is required for cross-century date accuracy. If year is between 0 and 99 is used, then year is assumed to be 1900 + year.
$ * @param monthIndex The month as a number between 1 and 12 (January to December).
$ * @param date The date as a number between 1 and 31.
$ * @param hours Must be supplied if minutes is supplied. A number from 0 to 23 (midnight to 11pm) that specifies the hour.
$ * @param minutes Must be supplied if seconds is supplied. A number from 0 to 59 that specifies the minutes.
$ * @param seconds Must be supplied if milliseconds is supplied. A number from 0 to 59 that specifies the seconds.
$ * @param ms A number from 0 to 999 that specifies the milliseconds.
class Date {
  private year = 0;
  private month = 0;
  private date = 0;
  private week = 0;
  private second = 0;
  $ * Raw(Time)
  private default time = null;

  constructor(year, monthIndex, date, hours, minutes, seconds, ms) {
    switch (typeof year) {
      case "number": {
        if (typeof monthIndex == "number") {
          if (typeof date != "number" || typeof hours != "number" || typeof minutes != "number" || typeof seconds != "number" || typeof ms != "number") throw #_new_error("Date: expects arguments of types (number)");
          this.time = #_time(year, monthIndex, date, hours, minutes, seconds, ms);
        }
        else {
          throw "Unimplemented.";
        }
      }
      case "string": this.time = #_time(year);
      default: this.time = #_time();
    }
  }

  $ *
  $ * (): number;
  public getTime() {
    return #_unix_milli(this.time);
  }

  $ *
  $ * (): number;
  public getYear() {
    return #_get_year(this.time);
  }

  $ *
  $ * (): string;
  public getMonth() {
    return #_get_month(this.time);
  }

  $ *
  $ * (): string;
  public getMonthString() {
    return Months[#_get_month(this.time) - 1];
  }

  $ *
  $ * (): number;
  public getDate() {
    return #_get_date(this.time);
  }

  $ *
  $ * (): string;
  public getDay() {
    return #_get_weekday(this.time);
  }

  $ *
  $ * (): number;
  public getHours() {
    return #_get_hour(this.time);
  }

  $ *
  $ * (): number;
  public getMinutes() {
    return #_get_minute(this.time);
  }

  $ *
  $ * (): number;
  public getSeconds() {
    return #_get_second(this.time);
  }

  $ *
  $ * (): number;
  public getMillisecond() {
    return #_get_millisec(this.time);
  }

  public toDateString() {
    return this.getDay() + " " + this.getMonthString() + " " + this.getDate() + " " + this.getYear();
  }

  public toTimeString() {
    return String(this.getHours()).padStart(2, '0') + ":" + String(this.getMinutes()).padStart(2, '0') + ":" + String(this.getSeconds()).padStart(2, '0') + " " $ + #_get_time_loc(this.time);
  }

  public toString() {
    return this.toDateString() + " " + this.toTimeString();
  }

  private [Symbol.toPrimitive]() {
    return this.toString();
  }

  private [Symbol.debug]() {
    return "\x1b[35m"+String(this.getYear()).padStart(4, '0') + "-" + String(this.getMonth()).padStart(2, '0') + "-" + String(this.getDate()).padStart(2, '0') + " " + this.toTimeString()+"\x1b[0m";
  }
}

Date.now = function() {
  return #_unix_milli(#_time())
}

Date.parse = function(str) {
  if (typeof str != "string") throw #_new_error("Date.parse: argument must be a string.");
  return #_unix_milli(#_time(str));
}