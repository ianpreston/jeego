**jeego** is a Twilio clone built on FreeSWITCH.

## Install

Clone the git repository

    git clone https://github.com/ianpreston/jeego.git

Build the source

    go build

Create a config file

    cp config-example.xml /etc/jeego.xml

Add a Dialplan entry to FreeSWITCH

    <extension name="jeego">
      <condition field="destination_number" expression="^2600$">
        <action application="socket" data="127.0.0.1:8084 full"/>
      </condition>
    </extension>

Run

	./jeego

## License

Created by [Ian Preston](https://ian-preston.com).

Available under the MIT License.