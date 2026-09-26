package com.fileshare.app

import android.Manifest
import android.annotation.SuppressLint
import android.content.pm.PackageManager
import android.os.Build
import android.os.Bundle
import android.os.Environment
import android.webkit.WebSettings
import android.webkit.WebView
import android.webkit.WebViewClient
import android.widget.Toast
import androidx.appcompat.app.AppCompatActivity
import androidx.core.app.ActivityCompat
import androidx.core.content.ContextCompat
import mobile.Mobile
import java.io.File

class MainActivity : AppCompatActivity() {

    private lateinit var webView: WebView
    private val PORT = 8990
    private val PERMISSION_REQUEST_CODE = 100

    @SuppressLint("SetJavaScriptEnabled")
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        
        webView = WebView(this)
        setContentView(webView)

        webView.webViewClient = WebViewClient()
        val webSettings: WebSettings = webView.settings
        webSettings.javaScriptEnabled = true
        webSettings.domStorageEnabled = true
        
        // No user tracking, cache is kept locally only for UI
        webSettings.cacheMode = WebSettings.LOAD_DEFAULT

        checkPermissionsAndStartServer()
    }

    private fun checkPermissionsAndStartServer() {
        val permissions = mutableListOf(
            Manifest.permission.INTERNET,
            Manifest.permission.ACCESS_NETWORK_STATE,
            Manifest.permission.ACCESS_WIFI_STATE,
            Manifest.permission.CHANGE_WIFI_MULTICAST_STATE
        )

        if (Build.VERSION.SDK_INT <= Build.VERSION_CODES.P) {
            permissions.add(Manifest.permission.WRITE_EXTERNAL_STORAGE)
        }

        val neededPermissions = permissions.filter {
            ContextCompat.checkSelfPermission(this, it) != PackageManager.PERMISSION_GRANTED
        }

        if (neededPermissions.isNotEmpty()) {
            ActivityCompat.requestPermissions(
                this,
                neededPermissions.toTypedArray(),
                PERMISSION_REQUEST_CODE
            )
        } else {
            startGoServer()
        }
    }

    override fun onRequestPermissionsResult(
        requestCode: Int,
        permissions: Array<out String>,
        grantResults: IntArray
    ) {
        super.onRequestPermissionsResult(requestCode, permissions, grantResults)
        if (requestCode == PERMISSION_REQUEST_CODE) {
            // Start server regardless of storage permission for discovery,
            // though without storage downloading will fail.
            startGoServer()
        }
    }

    private fun startGoServer() {
        try {
            // As per user constraint: save strictly in Downloads/FileShare
            val downloadsFolder = Environment.getExternalStoragePublicDirectory(Environment.DIRECTORY_DOWNLOADS)
            val fileShareDir = File(downloadsFolder, "FileShare")
            if (!fileShareDir.exists()) {
                fileShareDir.mkdirs()
            }

            // Start the Go Backend via gomobile binding
            Mobile.startServer(PORT.toLong(), fileShareDir.absolutePath)
            
            // Load the locally served Android-specific Web UI
            webView.loadUrl("http://localhost:$PORT/android.html")
            
        } catch (e: Exception) {
            Toast.makeText(this, "Failed to start FileShare Server: ${e.message}", Toast.LENGTH_LONG).show()
        }
    }

    override fun onDestroy() {
        super.onDestroy()
        try {
            Mobile.stopServer()
        } catch (e: Exception) {
            e.printStackTrace()
        }
    }
}
