import QtQuick 2.12
import QtQuick.Window 2.12
import QtQuick.Layouts 1.3
import QtQuick.Controls 2.13
import QtQuick.Dialogs.qml 1.0

Window {
    visible: true
    width: 640
    height: 400
    title: qsTr("Hello World")

    RowLayout {
        id: rowLayout
        width: 642
        anchors.fill: parent
        RoundButton {
            id: closeButton
            x: 0
            y: 0
            width: 10
            height: 10
            text: ""
            clip: false
            Layout.alignment: Qt.AlignLeft | Qt.AlignTop
            autoRepeat: false
            checkable: false
            checked: false
            display: AbstractButton.IconOnly
            transformOrigin: Item.TopLeft
            focusPolicy: Qt.StrongFocus
        }

        RoundButton {
            id: minButton
            x: 40
            y: 0
            width: 10
            height: 10
            text: ""
            spacing: 4
            focusPolicy: Qt.NoFocus
            autoRepeat: false
            checkable: false
            clip: false
            checked: false
            Layout.alignment: Qt.AlignLeft | Qt.AlignTop
            display: AbstractButton.IconOnly
            transformOrigin: Item.TopLeft
        }
        RoundButton {
            id: maxButton
            x: 80
            y: 0
            width: 10
            height: 10
            text: ""
            spacing: 4
            focusPolicy: Qt.NoFocus
            autoRepeat: false
            checkable: false
            clip: false
            checked: false
            Layout.alignment: Qt.AlignLeft | Qt.AlignTop
            display: AbstractButton.IconOnly
            transformOrigin: Item.TopLeft
        }
    }


}
